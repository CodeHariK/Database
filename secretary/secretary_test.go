package secretary

import "testing"

func TestTree(t *testing.T) {
	usersTree, userErr := NewBTree(
		"users",
		10,
		32,
		1024,
		125,
		10,
	)
	imagesTree, imagesErr := NewBTree(
		"images",
		100,
		32,
		1024*1024,
		125,
		10,
	)
	if userErr != nil || imagesErr != nil {
		t.Fatal(userErr, imagesErr)
	}

	usersTree.SaveHeader()
	imagesTree.SaveHeader()
}
