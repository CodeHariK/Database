package secretary

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/codeharik/secretary/utils"
)

var COMMAND_LOGS = ""

func ServerLog(msgs ...any) {
	if !MODE_TEST {
		msg, _ := utils.LogMessage(msgs...)
		COMMAND_LOGS += fmt.Sprintf("<div style='color:%s;background:#000'>%s</div><br>", utils.LightColor().Hex, strings.ReplaceAll(msg, "\n", "<br>"))
	}
}

type JsonResponse struct {
	Data any    `json:"data"`
	Logs string `json:"logs"`
}

func makeJson(data any) ([]byte, error) {
	response := JsonResponse{
		Data: data,
		Logs: COMMAND_LOGS,
	}

	return json.Marshal(response)
}

func (s *Secretary) HandleGetAllTree() ([]byte, error) {
	var trees []*BTree
	for _, o := range s.trees {
		trees = append(trees, o)
	}

	return makeJson(trees)
}

func (s *Secretary) HandleGetTree(collectionName string) ([]byte, error) {
	tree, exists := s.trees[collectionName]
	if !exists {
		return nil, ErrorTreeNotFound
	}

	jsonData, err := makeJson(tree.ToJSON())
	if err != nil {
		return nil, err
	}

	errs := tree.TreeVerify()
	return jsonData, errors.Join(errs...)
}

func (s *Secretary) HandleNewTree(collectionName string, order int, numLevel int, baseSize int, increment int, compactionBatchSize int) ([]byte, error) {
	tree, err := s.NewBTree(
		collectionName,
		uint8(order),
		uint8(numLevel),
		uint32(baseSize),
		uint8(increment),
		uint32(compactionBatchSize),
	)
	if err != nil {
		return nil, err
	}
	err = tree.SaveHeader()
	if err != nil {
		return nil, err
	}
	return makeJson("New tree created")
}

func (s *Secretary) HandleSetRecord(collectionName string, reqKey string, reqValue string) (data []byte, err error) {
	tree, exists := s.trees[collectionName]
	if !exists {
		return nil, ErrorTreeNotFound
	}

	key := []byte(reqKey)

	if len(reqKey) == 0 || len(reqKey) != KEY_SIZE {
		key = []byte(utils.GenerateSeqString(&tree.KeySeq, KEY_SIZE, KEY_INCREMENT))
		_, err = tree.SetKV(key, []byte(reqValue))
	} else {
		_, err = tree.SetKV(key, []byte(reqValue))
	}

	if err == ErrorDuplicateKey {
		err := tree.Update(key, []byte(reqValue))
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	if errs := tree.TreeVerify(); len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	response := map[string]any{
		"message":        "Data set successfully",
		"collectionName": collectionName,
		"key":            key,
	}

	data, err = makeJson(response)
	return data, err
}

func (s *Secretary) HandleSortedSetRecord(collectionName string, value int) ([]byte, error) {
	tree, exists := s.trees[collectionName]
	if !exists {
		return nil, ErrorTreeNotFound
	}

	tree.Erase()

	sortedRecords := SampleSortedKeyRecords(value)

	tree.SortedRecordSet(sortedRecords)

	if errs := tree.TreeVerify(); len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	response := map[string]any{
		"message":        "Data set successfully",
		"collectionName": collectionName,
	}

	return makeJson(response)
}

func (s *Secretary) HandleGetRecord(collectionName string, key string) ([]byte, error) {
	tree, exists := s.trees[collectionName]
	if !exists {
		return nil, ErrorTreeNotFound
	}

	node, index, found := tree.getLeafNode([]byte(key))
	if found {
		response := map[string]any{
			"collectionName": collectionName,
			"nodeID":         node.NodeID,
			"found":          found,
			"record":         node.records[index].Value,
		}
		return makeJson(response)
	}

	return nil, ErrorKeyNotFound
}

func (s *Secretary) HandleDeleteRecord(collectionName string, id string) ([]byte, error) {
	tree, exists := s.trees[collectionName]
	if !exists {
		return nil, ErrorTreeNotFound
	}

	err := tree.Delete([]byte(id))
	if err != nil {
		return nil, err
	}
	if errs := tree.TreeVerify(); len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	response := map[string]any{
		"collectionName": collectionName,
		"result":         "Delete success " + id,
	}

	return makeJson(response)
}

func (s *Secretary) HandleClearTree(collectionName string) ([]byte, error) {
	tree, exists := s.trees[collectionName]
	if !exists {
		return nil, ErrorTreeNotFound
	}

	tree.Erase()

	response := map[string]any{
		"collectionName": collectionName,
		"result":         "Clear tree success",
	}

	return makeJson(response)
}
