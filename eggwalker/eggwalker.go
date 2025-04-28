package eggwalker

import "errors"

type Id struct {
	agent string
	seq   int32
}

func (b *Id) equals(a *Id) bool {
	return a != nil && b != nil && a.agent == b.agent && a.seq == b.seq
}

type Item struct {
	content byte

	id          *Id
	originLeft  *Id
	originRight *Id

	deleted bool
}

type Doc struct {
	content []Item
}

func createDoc() Doc {
	return Doc{
		content: []Item{},
	}
}

func getContent(doc Doc) string {
	notDeleted := make([]byte, len(doc.content))
	i := 0
	for _, item := range doc.content {
		if !item.deleted {
			notDeleted[i] = item.content
			i++
		}
	}
	return string(notDeleted[:i])
}

func (doc Doc) localInsertOne(
	agent string,
	seq int32,
	pos int32,
	text byte,
) {
	var originLeft *Id
	if ((pos - 1) > 0) && ((pos - 1) < int32(len(doc.content))) {
		originLeft = doc.content[pos-1].id
	}
	var originRight *Id
	if (pos > 0) && (pos < int32(len(doc.content))) {
		originRight = doc.content[pos].id
	}
	doc.intergrate(Item{
		content: text,
		id: &Id{
			agent: agent,
			seq:   seq,
		},
		deleted:     false,
		originLeft:  originLeft,
		originRight: originRight,
	})
}

func findItemIdxAtId(doc Doc, id *Id) (int32, error) {
	if id == nil {
		return -1, errors.New("Invalid id")
	}

	for i := int32(0); int(i) < len(doc.content); i++ {
		if id.equals(doc.content[i].id) {
			return i, nil
		}
	}

	return -1, errors.New("Item not found")
}

func (doc Doc) intergrate(newItem Item) {
	// Add newItem into doc at right location
}
