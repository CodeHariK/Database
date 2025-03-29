package secretary

import (
	"testing"

	"github.com/codeharik/secretary/utils/binstruct"
)

func TestSecretary(t *testing.T) {
	s := dummySecretary(t)
	usersTree, userErr := s.NewBTree(
		"users",
		4,
		32,
		1024,
		125,
		1000,
	)

	imagesTree, imagesErr := s.NewBTree(
		"images",
		4,
		32,
		1024*1024,
		125,
		1000,
	)
	if userErr != nil || imagesErr != nil {
		t.Fatal(userErr, imagesErr)
	}

	userErr = usersTree.SaveHeader()
	imagesErr = imagesTree.SaveHeader()
	if userErr != nil || imagesErr != nil {
		t.Fatal(userErr, imagesErr)
	}

	newSecretary := dummySecretary(t)

	users, err := newSecretary.Tree("users")
	if err != nil {
		t.Fatal(err)
	}
	images, err := newSecretary.Tree("images")
	if err != nil {
		t.Fatal(err)
	}

	eq, err := binstruct.Compare(usersTree, users)
	if !eq || err != nil {
		t.Fatalf("Should be equal %+v : %+v", usersTree, users)
	}
	eq, err = binstruct.Compare(imagesTree, images)
	if !eq || err != nil {
		t.Fatalf("Should be equal %+v : %+v", imagesTree, images)
	}

	_, err = newSecretary.Tree("unknown")
	if err == nil {
		t.Fatal(err)
	}

	s.PagerShutdown()
}
