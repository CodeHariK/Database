package secretary

import (
	"testing"

	"github.com/codeharik/secretary/utils/binstruct"
)

func TestSecretary(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	usersTree, userErr := s.NewBTree(
		"users",
		4,
		32,
		1024,
		125,
		10,
	)

	imagesTree, imagesErr := s.NewBTree(
		"images",
		4,
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

	newSecretary, err := New()
	if err != nil {
		t.Fatal(err)
	}

	users, err := newSecretary.Tree("users")
	if err != nil {
		t.Fatal(err)
	}
	images, err := newSecretary.Tree("images")
	if err != nil {
		t.Fatal(err)
	}

	eq, err := binstruct.Compare(*usersTree, *users)
	if !eq || err != nil {
		t.Fatalf("Should be equal %+v : %+v", *usersTree, *users)
	}
	eq, err = binstruct.Compare(*imagesTree, *images)
	if !eq || err != nil {
		t.Fatalf("Should be equal %+v : %+v", *imagesTree, *images)
	}

	_, err = newSecretary.Tree("unknown")
	if err == nil {
		t.Fatal(err)
	}
}
