{{define "schemaVar"}}pkgSchema.{{.Table|upperCamelCase}}Schema{{end}}
{{define "paramList"}}{{range $idx, $pk := .}}{{if $idx}}, {{end}}{{$pk.ColumnName|lowerCamelCase}} {{$pk.Type}}{{end}}{{end}}
{{define "argList"}}{{range $idx, $pk := .}}{{if $idx}}, {{end}}{{$pk.ColumnName|lowerCamelCase}}{{end}}{{end}}
{{define "whereMatch"}}{{range $idx, $pk := .}}{{if $idx}} AND {{end}}{{$pk.ColumnName}} = ${{add $idx 1}}{{end}}{{end}}
{{define "commaSeparatedColumns"}}{{range $idx, $field := .}}{{if $idx}}, {{end}}{{$field.ColumnName}}{{end}}{{end}}
{{define "commandSeparatedRefs"}}{{range $idx, $field := .}}{{if $idx}}, {{end}}{{$field.Reference}}{{end}}{{end}}
{{define "updateExclusions"}}{{range $idx, $field := .}}{{if $idx}}, {{end}}{{$field.ColumnName}} = EXCLUDED.{{$field.ColumnName}}{{end}}{{end}}

{{- $ := . }}
{{- $pks := .Schema.PrimaryKeys }}

{{- $singlePK := false }}
{{- if eq (len $pks) 1 }}
{{ $singlePK = index $pks 0 }}
{{/*If there are multiple pks, then use the explicitly specified id column.*/}}
{{- else if .Schema.ID.ColumnName}}
{{ $singlePK = .Schema.ID }}
{{- end }}

package postgres

import (
    "context"
    "strings"
    "time"

    "github.com/gogo/protobuf/proto"
    "github.com/jackc/pgx/v4"
    "github.com/jackc/pgx/v4/pgxpool"
    "github.com/pkg/errors"
    "github.com/stackrox/rox/central/metrics"
    pkgSchema "github.com/stackrox/rox/central/postgres/schema"
    "github.com/stackrox/rox/central/role/resources"
    v1 "github.com/stackrox/rox/generated/api/v1"
    "github.com/stackrox/rox/generated/storage"
    "github.com/stackrox/rox/pkg/auth/permissions"
    "github.com/stackrox/rox/pkg/logging"
    ops "github.com/stackrox/rox/pkg/metrics"
    "github.com/stackrox/rox/pkg/postgres/pgutils"
    "github.com/stackrox/rox/pkg/sac"
    "github.com/stackrox/rox/pkg/search/postgres"
)

const (
        baseTable = "{{.Table}}"
        existsStmt = "SELECT EXISTS(SELECT 1 FROM {{.Table}} WHERE {{template "whereMatch" $pks}})"

        getStmt = "SELECT serialized FROM {{.Table}} WHERE {{template "whereMatch" $pks}}"
        deleteStmt = "DELETE FROM {{.Table}} WHERE {{template "whereMatch" $pks}}"
        walkStmt = "SELECT serialized FROM {{.Table}}"

{{- if $singlePK }}
        getIDsStmt = "SELECT {{$singlePK.ColumnName}} FROM {{.Table}}"
        getManyStmt = "SELECT serialized FROM {{.Table}} WHERE {{$singlePK.ColumnName}} = ANY($1::text[])"

        deleteManyStmt = "DELETE FROM {{.Table}} WHERE {{$singlePK.ColumnName}} = ANY($1::text[])"
{{- end }}

        batchAfter = 100

        // using copyFrom, we may not even want to batch.  It would probably be simpler
        // to deal with failures if we just sent it all.  Something to think about as we
        // proceed and move into more e2e and larger performance testing
        batchSize = 10000
)

var (
    log = logging.LoggerForModule()
    schema = {{ template "schemaVar" .Schema}}
    {{ if or (.Obj.IsGloballyScoped) (.Obj.IsDirectlyScoped) -}}
    targetResource = resources.{{.Type | storageToResource}}
    {{- end }}
)

type Store interface {
    Count(ctx context.Context) (int, error)
    Exists(ctx context.Context, {{template "paramList" $pks}}) (bool, error)
    Get(ctx context.Context, {{template "paramList" $pks}}) (*{{.Type}}, bool, error)
{{- if .GetAll }}
    GetAll(ctx context.Context) ([]*{{.Type}}, error)
{{- end }}
{{- if not .JoinTable }}
    Upsert(ctx context.Context, obj *{{.Type}}) error
    UpsertMany(ctx context.Context, objs []*{{.Type}}) error
    Delete(ctx context.Context, {{template "paramList" $pks}}) error
{{- end }}

{{- if $singlePK }}
    GetIDs(ctx context.Context) ([]{{$singlePK.Type}}, error)
    GetMany(ctx context.Context, ids []{{$singlePK.Type}}) ([]*{{.Type}}, []int, error)
{{- if not .JoinTable }}
    DeleteMany(ctx context.Context, ids []{{$singlePK.Type}}) error
{{- end }}
{{- end }}

    Walk(ctx context.Context, fn func(obj *{{.Type}}) error) error

    AckKeysIndexed(ctx context.Context, keys ...string) error
    GetKeysToIndex(ctx context.Context) ([]string, error)
}

type storeImpl struct {
    db *pgxpool.Pool
}

{{ define "defineScopeChecker" }}scopeChecker := sac.GlobalAccessScopeChecker(ctx).AccessMode(storage.Access_{{ . }}_ACCESS).Resource(targetResource){{ end }}

{{define "createTableStmtVar"}}pkgSchema.CreateTable{{.Table|upperCamelCase}}Stmt{{end}}
{{- define "createTable" }}
{{- $schema := . }}
pgutils.CreateTable(ctx, db, {{template "createTableStmtVar" $schema}})
{{- end }}

// New returns a new Store instance using the provided sql instance.
func New(ctx context.Context, db *pgxpool.Pool) Store {
    {{- range $reference := .Schema.References }}
    {{- template "createTable" $reference.OtherSchema }}
    {{- end }}
    {{- template "createTable" .Schema}}

    return &storeImpl{
        db: db,
    }
}

{{- define "insertFunctionName"}}{{- $schema := . }}insertInto{{$schema.Table|upperCamelCase}}
{{- end}}

{{- define "insertObject"}}
{{- $schema := .schema }}
func {{ template "insertFunctionName" $schema }}(ctx context.Context, tx pgx.Tx, obj {{$schema.Type}}{{ range $field := $schema.FieldsDeterminedByParent }}, {{$field.Name}} {{$field.Type}}{{end}}) error {
    {{if not $schema.Parent }}
    serialized, marshalErr := obj.Marshal()
    if marshalErr != nil {
        return marshalErr
    }
    {{end}}

    values := []interface{} {
        // parent primary keys start
        {{- range $field := $schema.DBColumnFields -}}
        {{- if eq $field.DataType "datetime" }}
        pgutils.NilOrTime({{$field.Getter "obj"}}),
        {{- else }}
        {{$field.Getter "obj"}},{{end}}
        {{- end}}
    }

    finalStr := "INSERT INTO {{$schema.Table}} ({{template "commaSeparatedColumns" $schema.DBColumnFields }}) VALUES({{ valueExpansion (len $schema.DBColumnFields) }}) ON CONFLICT({{template "commaSeparatedColumns" $schema.PrimaryKeys}}) DO UPDATE SET {{template "updateExclusions" $schema.DBColumnFields}}"
    _, err := tx.Exec(ctx, finalStr, values...)
    if err != nil {
        return err
    }

    {{ if $schema.Children }}
    var query string
    {{end}}

    {{range $idx, $child := $schema.Children }}
    for childIdx, child := range obj.{{$child.ObjectGetter}} {
        if err := {{ template "insertFunctionName" $child }}(ctx, tx, child{{ range $field := $schema.PrimaryKeys }}, {{$field.Getter "obj"}}{{end}}, childIdx); err != nil {
            return err
        }
    }

    query = "delete from {{$child.Table}} where {{ range $idx, $field := $child.FieldsReferringToParent }}{{if $idx}} AND {{end}}{{$field.ColumnName}} = ${{add $idx 1}}{{end}} AND idx >= ${{add (len $child.FieldsReferringToParent) 1}}"
    _, err = tx.Exec(ctx, query{{ range $field := $schema.PrimaryKeys }}, {{$field.Getter "obj"}}{{end}}, len(obj.{{$child.ObjectGetter}}))
    if err != nil {
        return err
    }

    {{- end}}
    return nil
}

{{range $idx, $child := $schema.Children}}{{ template "insertObject" dict "schema" $child "joinTable" false }}{{end}}
{{- end}}

{{- if not .JoinTable }}
{{ template "insertObject" dict "schema" .Schema "joinTable" .JoinTable }}
{{- end}}

{{- define "copyFunctionName"}}{{- $schema := . }}copyFrom{{$schema.Table|upperCamelCase}}
{{- end}}

{{- define "copyObject"}}
{{- $schema := . }}
func (s *storeImpl) {{ template "copyFunctionName" $schema }}(ctx context.Context, tx pgx.Tx, {{ range $idx, $field := $schema.FieldsReferringToParent }} {{$field.Name}} {{$field.Type}},{{end}} objs ...{{$schema.Type}}) error {

    inputRows := [][]interface{}{}

    var err error

    {{if and (eq (len $schema.PrimaryKeys) 1) (not $schema.Parent) }}
    // This is a copy so first we must delete the rows and re-add them
    // Which is essentially the desired behaviour of an upsert.
    var deletes []string
    {{end}}

    copyCols := []string {
    {{range $idx, $field := $schema.DBColumnFields}}
        "{{$field.ColumnName|lowerCase}}",
    {{end}}
    }

    for idx, obj := range objs {
        // Todo: ROX-9499 Figure out how to more cleanly template around this issue.
        log.Debugf("This is here for now because there is an issue with pods_TerminatedInstances where the obj in the loop is not used as it only consists of the parent id and the idx.  Putting this here as a stop gap to simply use the object.  %s", obj)

        {{/* If embedded, the top-level has the full serialized object */}}
        {{if not $schema.Parent }}
        serialized, marshalErr := obj.Marshal()
        if marshalErr != nil {
            return marshalErr
        }
        {{end}}

        inputRows = append(inputRows, []interface{}{
            {{ range $idx, $field := $schema.DBColumnFields }}
            {{if eq $field.DataType "datetime"}}
            pgutils.NilOrTime({{$field.Getter "obj"}}),
            {{- else}}
            {{$field.Getter "obj"}},{{end}}
            {{end}}
        })

        {{ if not $schema.Parent }}
        {{if eq (len $schema.PrimaryKeys) 1}}
        // Add the id to be deleted.
        deletes = append(deletes, {{ range $field := $schema.PrimaryKeys }}{{$field.Getter "obj"}}, {{end}})
        {{else}}
        if _, err := tx.Exec(ctx, deleteStmt, {{ range $field := $schema.PrimaryKeys }}{{$field.Getter "obj"}}, {{end}}); err != nil {
            return err
        }

        {{end}}
        {{end}}

        // if we hit our batch size we need to push the data
        if (idx + 1) % batchSize == 0 || idx == len(objs) - 1  {
            // copy does not upsert so have to delete first.  parent deletion cascades so only need to
            // delete for the top level parent
            {{if and ((eq (len $schema.PrimaryKeys) 1)) (not $schema.Parent) }}
            _, err = tx.Exec(ctx, deleteManyStmt, deletes);
            if err != nil {
                return err
            }
            // clear the inserts and vals for the next batch
            deletes = nil
            {{end}}

            _, err = tx.CopyFrom(ctx, pgx.Identifier{"{{$schema.Table|lowerCase}}"}, copyCols, pgx.CopyFromRows(inputRows))

            if err != nil {
                return err
            }

            // clear the input rows for the next batch
            inputRows = inputRows[:0]
        }
    }

    {{if $schema.Children }}
    for idx, obj := range objs {
        _ = idx // idx may or may not be used depending on how nested we are, so avoid compile-time errors.
        {{range $child := $schema.Children }}
        if err = s.{{ template "copyFunctionName" $child }}(ctx, tx{{ range $idx, $field := $schema.PrimaryKeys }}, {{$field.Getter "obj"}}{{end}}, obj.{{$child.ObjectGetter}}...); err != nil {
            return err
        }
        {{- end}}
    }
    {{end}}

    return err
}
{{range $child := $schema.Children}}{{ template "copyObject" $child }}{{end}}
{{- end}}

{{- if not .JoinTable }}
{{ template "copyObject" .Schema }}
{{- end }}

{{- if not .JoinTable }}

func (s *storeImpl) copyFrom(ctx context.Context, objs ...*{{.Type}}) error {
    conn, release, err := s.acquireConn(ctx, ops.Get, "{{.TrimmedType}}")
	if err != nil {
	    return err
	}
    defer release()

    tx, err := conn.Begin(ctx)
    if err != nil {
        return err
    }

    if err := s.{{ template "copyFunctionName" .Schema }}(ctx, tx, objs...); err != nil {
        if err := tx.Rollback(ctx); err != nil {
            return err
        }
        return err
    }
    if err := tx.Commit(ctx); err != nil {
        return err
    }
    return nil
}

func (s *storeImpl) upsert(ctx context.Context, objs ...*{{.Type}}) error {
    conn, release, err := s.acquireConn(ctx, ops.Get, "{{.TrimmedType}}")
	if err != nil {
	    return err
	}
	defer release()

    for _, obj := range objs {
	    tx, err := conn.Begin(ctx)
	    if err != nil {
    		return err
	    }

	    if err := {{ template "insertFunctionName" .Schema }}(ctx, tx, obj); err != nil {
		    if err := tx.Rollback(ctx); err != nil {
			    return err
		    }
		    return err
        }
        if err := tx.Commit(ctx); err != nil {
            return err
        }
    }
    return nil
}

func (s *storeImpl) Upsert(ctx context.Context, obj *{{.Type}}) error {
    defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.Upsert, "{{.TrimmedType}}")

    {{ if .PermissionChecker -}}
    if ok, err := {{ .PermissionChecker }}.UpsertAllowed(ctx); err != nil {
        return err
    } else if !ok {
        return sac.ErrResourceAccessDenied
    }
    {{- else if .Obj.IsGloballyScoped }}
    {{ template "defineScopeChecker" "READ_WRITE" }}
    {{- else if and (.Obj.IsDirectlyScoped) (.Obj.IsClusterScope) }}
    {{ template "defineScopeChecker" "READ_WRITE" }}.
        ClusterID({{ "obj" | .Obj.GetClusterID }})
    {{- else if and (.Obj.IsDirectlyScoped) (.Obj.IsNamespaceScope) }}
    {{ template "defineScopeChecker" "READ_WRITE" }}.
        ClusterID({{ "obj" | .Obj.GetClusterID }}).Namespace({{ "obj" | .Obj.GetNamespace }})
    {{- end }}
    {{- if or (.Obj.IsGloballyScoped) (.Obj.IsDirectlyScoped) }}
    if ok, err := scopeChecker.Allowed(ctx); err != nil {
        return err
    } else if !ok {
        return sac.ErrResourceAccessDenied
    }
    {{- end }}

    return s.upsert(ctx, obj)
}

func (s *storeImpl) UpsertMany(ctx context.Context, objs []*{{.Type}}) error {
    defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.UpdateMany, "{{.TrimmedType}}")

    {{ if .PermissionChecker -}}
    if ok, err := {{ .PermissionChecker }}.UpsertManyAllowed(ctx); err != nil {
        return err
    } else if !ok {
        return sac.ErrResourceAccessDenied
    }
    {{- else if .Obj.IsGloballyScoped }}
    {{ template "defineScopeChecker" "READ_WRITE" }}
    if ok, err := scopeChecker.Allowed(ctx); err != nil {
        return err
    } else if !ok {
        return sac.ErrResourceAccessDenied
    }
    {{- else if .Obj.IsDirectlyScoped -}}
    {{ template "defineScopeChecker" "READ_WRITE" }}
    if ok, err := scopeChecker.Allowed(ctx); err != nil {
        return err
    } else if !ok {
        var deniedIds []string
        for _, obj := range objs {
            {{- if .Obj.IsClusterScope }}
            subScopeChecker := scopeChecker.ClusterID({{ "obj" | .Obj.GetClusterID }})
            {{- else if .Obj.IsNamespaceScope }}
            subScopeChecker := scopeChecker.ClusterID({{ "obj" | .Obj.GetClusterID }}).Namespace({{ "obj" | .Obj.GetNamespace }})
            {{- end }}
            if ok, err := subScopeChecker.Allowed(ctx); err != nil {
                return err
            } else if !ok {
                deniedIds = append(deniedIds, obj.GetId())
            }
        }
        if len(deniedIds) != 0 {
            return errors.Wrapf(sac.ErrResourceAccessDenied, "modifying {{ .TrimmedType|lowerCamelCase }}s with IDs [%s] was denied", strings.Join(deniedIds, ", "))
        }
    }
    {{- end }}

    if len(objs) < batchAfter {
        return s.upsert(ctx, objs...)
    } else {
        return s.copyFrom(ctx, objs...)
    }
}
{{- end }}

// Count returns the number of objects in the store
func (s *storeImpl) Count(ctx context.Context) (int, error) {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.Count, "{{.TrimmedType}}")

    var sacQueryFilter *v1.Query

    {{ if .PermissionChecker -}}
    if ok, err := {{ .PermissionChecker }}.CountAllowed(ctx); err != nil || !ok {
        return 0, err
    }
    {{- else if .Obj.IsGloballyScoped }}
    {{ template "defineScopeChecker" "READ" }}
    if ok, err := scopeChecker.Allowed(ctx); err != nil || !ok {
        return 0, err
    }
    {{- else if .Obj.IsDirectlyScoped }}
    scopeChecker := sac.GlobalAccessScopeChecker(ctx)
	scopeTree, err := scopeChecker.EffectiveAccessScope(permissions.ResourceWithAccess{
		Resource: targetResource,
		Access:   storage.Access_READ_ACCESS,
	})
	if err != nil {
		return 0, err
	}
    {{- if .Obj.IsClusterScope }}
    sacQueryFilter, err = sac.BuildClusterLevelSACQueryFilter(scopeTree)
    {{- else}}
    sacQueryFilter, err = sac.BuildClusterNamespaceLevelSACQueryFilter(scopeTree)
    {{- end }}

	if err != nil {
		return 0, err
	}
    {{- end }}

    return postgres.RunCountRequestForSchema(schema, sacQueryFilter, s.db)
}

// Exists returns if the id exists in the store
func (s *storeImpl) Exists(ctx context.Context, {{template "paramList" $pks}}) (bool, error) {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.Exists, "{{.TrimmedType}}")

    {{ if .PermissionChecker -}}
    if ok, err := {{ .PermissionChecker }}.ExistsAllowed(ctx); err != nil || !ok {
        return false, err
    }
    {{- else if .Obj.IsGloballyScoped }}
    {{ template "defineScopeChecker" "READ" }}
    if ok, err := scopeChecker.Allowed(ctx); err != nil {
        return false, err
    } else if !ok {
        return false, nil
    }
    {{- end}}

	row := s.db.QueryRow(ctx, existsStmt, {{template "argList" $pks}})
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, pgutils.ErrNilIfNoRows(err)
	}
	return exists, nil
}

// Get returns the object, if it exists from the store
func (s *storeImpl) Get(ctx context.Context, {{template "paramList" $pks}}) (*{{.Type}}, bool, error) {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.Get, "{{.TrimmedType}}")

    {{ if .PermissionChecker -}}
    if ok, err := {{ .PermissionChecker }}.GetAllowed(ctx); err != nil || !ok {
        return nil, false, err
    }
    {{- else if .Obj.IsGloballyScoped }}
    {{ template "defineScopeChecker" "READ" }}
    if ok, err := scopeChecker.Allowed(ctx); err != nil {
        return nil, false, err
    } else if !ok {
        return nil, false, nil
    }
    {{- end}}

	conn, release, err := s.acquireConn(ctx, ops.Get, "{{.TrimmedType}}")
	if err != nil {
	    return nil, false, err
	}
	defer release()

	row := conn.QueryRow(ctx, getStmt, {{template "argList" $pks}})
	var data []byte
	if err := row.Scan(&data); err != nil {
		return nil, false, pgutils.ErrNilIfNoRows(err)
	}

	var msg {{.Type}}
	if err := proto.Unmarshal(data, &msg); err != nil {
        return nil, false, err
	}
	return &msg, true, nil
}

{{- if .GetAll }}
func(s *storeImpl) GetAll(ctx context.Context) ([]*{{.Type}}, error) {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.GetAll, "{{.TrimmedType}}")

    var objs []*{{.Type}}
    err := s.Walk(ctx, func(obj *{{.Type}}) error {
        objs = append(objs, obj)
        return nil
    })
    return objs, err
}
{{- end}}

func (s *storeImpl) acquireConn(ctx context.Context, op ops.Op, typ string) (*pgxpool.Conn, func(), error) {
	defer metrics.SetAcquireDBConnDuration(time.Now(), op, typ)
	conn, err := s.db.Acquire(ctx)
	if err != nil {
	    return nil, nil, err
	}
	return conn, conn.Release, nil
}

{{- if not .JoinTable }}
// Delete removes the specified ID from the store
func (s *storeImpl) Delete(ctx context.Context, {{template "paramList" $pks}}) error {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.Remove, "{{.TrimmedType}}")

	{{ if .PermissionChecker -}}
    if ok, err := {{ .PermissionChecker }}.DeleteAllowed(ctx); err != nil {
        return err
    } else if !ok {
        return sac.ErrResourceAccessDenied
    }
    {{- else if .Obj.IsGloballyScoped }}
    {{ template "defineScopeChecker" "READ_WRITE" }}
    if ok, err := scopeChecker.Allowed(ctx); err != nil {
        return err
    } else if !ok {
        return sac.ErrResourceAccessDenied
    }
    {{- end}}

    conn, release, err := s.acquireConn(ctx, ops.Remove, "{{.TrimmedType}}")
	if err != nil {
	    return err
	}
	defer release()

	if _, err := conn.Exec(ctx, deleteStmt, {{template "argList" $pks}}); err != nil {
		return err
	}
	return nil
}
{{- end}}

{{- if $singlePK }}

// GetIDs returns all the IDs for the store
func (s *storeImpl) GetIDs(ctx context.Context) ([]{{$singlePK.Type}}, error) {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.GetAll, "{{.Type}}IDs")

    {{ if .PermissionChecker -}}
    if ok, err := {{ .PermissionChecker }}.GetIDsAllowed(ctx); err != nil || !ok {
        return nil, err
    }
    {{- else if .Obj.IsGloballyScoped }}
    {{ template "defineScopeChecker" "READ" }}
    if ok, err := scopeChecker.Allowed(ctx); err != nil {
        return nil, err
    } else if !ok {
        return nil, nil
    }
    {{- end}}

	rows, err := s.db.Query(ctx, getIDsStmt)
	if err != nil {
		return nil, pgutils.ErrNilIfNoRows(err)
	}
	defer rows.Close()
	var ids []{{$singlePK.Type}}
	for rows.Next() {
		var id {{$singlePK.Type}}
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// GetMany returns the objects specified by the IDs or the index in the missing indices slice
func (s *storeImpl) GetMany(ctx context.Context, ids []{{$singlePK.Type}}) ([]*{{.Type}}, []int, error) {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.GetMany, "{{.TrimmedType}}")

    {{ if .Obj.HasPermissionChecker -}}
    if ok, err := {{ .PermissionChecker }}.GetManyAllowed(ctx); err != nil {
        return nil, nil, err
    } else if !ok {
        return nil, nil, nil
    }
    {{- else if .Obj.IsGloballyScoped }}
    {{ template "defineScopeChecker" "READ" }}
    if ok, err := scopeChecker.Allowed(ctx); err != nil {
        return nil, nil, err
    } else if !ok {
        return nil, nil, nil
    }
    {{- end}}

	conn, release, err := s.acquireConn(ctx, ops.GetMany, "{{.TrimmedType}}")
	if err != nil {
	    return nil, nil, err
	}
	defer release()

	rows, err := conn.Query(ctx, getManyStmt, ids)
	if err != nil {
		if err == pgx.ErrNoRows {
			missingIndices := make([]int, 0, len(ids))
			for i := range ids {
				missingIndices = append(missingIndices, i)
			}
			return nil, missingIndices, nil
		}
		return nil, nil, err
	}
	defer rows.Close()
	resultsByID := make(map[{{$singlePK.Type}}]*{{.Type}})
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, nil, err
		}
		msg := &{{.Type}}{}
		if err := proto.Unmarshal(data, msg); err != nil {
		    return nil, nil, err
		}
		resultsByID[{{$singlePK.Getter "msg"}}] = msg
	}
	missingIndices := make([]int, 0, len(ids)-len(resultsByID))
	// It is important that the elems are populated in the same order as the input ids
	// slice, since some calling code relies on that to maintain order.
	elems := make([]*{{.Type}}, 0, len(resultsByID))
	for i, id := range ids {
		if result, ok := resultsByID[id]; !ok {
			missingIndices = append(missingIndices, i)
		} else {
		    elems = append(elems, result)
		}
	}
	return elems, missingIndices, nil
}

{{- if not .JoinTable }}
// Delete removes the specified IDs from the store
func (s *storeImpl) DeleteMany(ctx context.Context, ids []{{$singlePK.Type}}) error {
	defer metrics.SetPostgresOperationDurationTime(time.Now(), ops.RemoveMany, "{{.TrimmedType}}")

    {{ if .PermissionChecker -}}
    if ok, err := {{ .PermissionChecker }}.DeleteManyAllowed(ctx); err != nil {
        return err
    } else if !ok {
        return sac.ErrResourceAccessDenied
    }
    {{- else if .Obj.IsGloballyScoped }}
    {{ template "defineScopeChecker" "READ_WRITE" }}
    if ok, err := scopeChecker.Allowed(ctx); err != nil {
        return err
    } else if !ok {
        return sac.ErrResourceAccessDenied
    }
    {{- end}}

	conn, release, err := s.acquireConn(ctx, ops.RemoveMany, "{{.TrimmedType}}")
	if err != nil {
	    return err
	}
	defer release()
	if _, err := conn.Exec(ctx, deleteManyStmt, ids); err != nil {
		return err
	}
	return nil
}
{{- end }}
{{- end }}

// Walk iterates over all of the objects in the store and applies the closure
func (s *storeImpl) Walk(ctx context.Context, fn func(obj *{{.Type}}) error) error {
	rows, err := s.db.Query(ctx, walkStmt)
	if err != nil {
		return pgutils.ErrNilIfNoRows(err)
	}
	defer rows.Close()
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return err
		}
		var msg {{.Type}}
		if err := proto.Unmarshal(data, &msg); err != nil {
		    return err
		}
		if err := fn(&msg); err != nil {
		    return err
		}
	}
	return nil
}

//// Used for testing
{{- define "dropTableFunctionName"}}dropTable{{.Table | upperCamelCase}}{{end}}
{{- define "dropTable"}}
{{- $schema := . }}
func {{ template "dropTableFunctionName" $schema }}(ctx context.Context, db *pgxpool.Pool) {
    _, _ = db.Exec(ctx, "DROP TABLE IF EXISTS {{$schema.Table}} CASCADE")
    {{range $child := $schema.Children}}{{ template "dropTableFunctionName" $child }}(ctx, db)
    {{end}}
}
{{range $child := $schema.Children}}{{ template "dropTable" $child }}{{end}}
{{- end}}

{{template "dropTable" .Schema}}

func Destroy(ctx context.Context, db *pgxpool.Pool) {
    {{template "dropTableFunctionName" .Schema}}(ctx, db)
}

//// Stubs for satisfying legacy interfaces

// AckKeysIndexed acknowledges the passed keys were indexed
func (s *storeImpl) AckKeysIndexed(ctx context.Context, keys ...string) error {
	return nil
}

// GetKeysToIndex returns the keys that need to be indexed
func (s *storeImpl) GetKeysToIndex(ctx context.Context) ([]string, error) {
	return nil, nil
}
