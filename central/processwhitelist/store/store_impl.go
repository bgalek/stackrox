// Code generated by boltbindings generator. DO NOT EDIT.

package store

import (
	bbolt "github.com/etcd-io/bbolt"
	proto1 "github.com/gogo/protobuf/proto"
	storage "github.com/stackrox/rox/generated/storage"
	bolthelper "github.com/stackrox/rox/pkg/bolthelper"
	proto "github.com/stackrox/rox/pkg/bolthelper/crud/proto"
)

var (
	bucketName = []byte("processWhitelists")
)

type store struct {
	crud proto.MessageCrud
}

func key(msg proto1.Message) []byte {
	return []byte(msg.(*storage.ProcessWhitelist).GetId())
}

func alloc() proto1.Message {
	return new(storage.ProcessWhitelist)
}

func newStore(db *bbolt.DB) (*store, error) {
	if err := bolthelper.RegisterBucket(db, bucketName); err != nil {
		return nil, err
	}
	return &store{crud: proto.NewMessageCrud(db, bucketName, key, alloc)}, nil
}

func (s *store) AddWhitelist(whitelist *storage.ProcessWhitelist) error {
	return s.crud.Create(whitelist)
}

func (s *store) DeleteWhitelist(id string) error {
	return s.crud.Delete(id)
}

func (s *store) GetWhitelist(id string) (*storage.ProcessWhitelist, error) {
	msg, err := s.crud.Read(id)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		return nil, nil
	}
	storedKey := msg.(*storage.ProcessWhitelist)
	return storedKey, nil
}

func (s *store) ListWhitelists() ([]*storage.ProcessWhitelist, error) {
	msgs, err := s.crud.ReadAll()
	if err != nil {
		return nil, err
	}
	storedKeys := make([]*storage.ProcessWhitelist, len(msgs))
	for i, msg := range msgs {
		storedKeys[i] = msg.(*storage.ProcessWhitelist)
	}
	return storedKeys, nil
}

func (s *store) UpdateWhitelist(whitelist *storage.ProcessWhitelist) error {
	return s.crud.Update(whitelist)
}
