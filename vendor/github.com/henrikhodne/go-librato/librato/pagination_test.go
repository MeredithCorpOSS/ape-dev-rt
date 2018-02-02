package librato

import (
	"testing"

	"github.com/google/go-querystring/query"
)

func TestEmbeddedPaginationIsNotNestedInQueryStringParameters(t *testing.T) {
	var options = struct{ *PaginationMeta }{
		PaginationMeta: &PaginationMeta{
			Offset:  10,
			Length:  20,
			OrderBy: "name",
			Sort:    "desc",
		},
	}

	expected := "length=20&offset=10&orderby=name&sort=desc"

	qs, err := query.Values(options)

	if err != nil {
		t.Fatalf("unexpected error when encoding values: %#v", err)
	}

	encoded := qs.Encode()
	if encoded != expected {
		t.Errorf("struct values were not extracted correctly: %#v", encoded)
	}

}

func TestEncodingOptionsStructWithoutPaginationDoesNotResultInPanic(t *testing.T) {
	var options = struct{ *PaginationMeta }{}

	qs, err := query.Values(options)

	if err != nil {
		t.Fatalf("unexpected error when encoding values: %#v", err)
	}

	encoded := qs.Encode()

	if encoded != "" {
		t.Errorf("expected encoded query string to be empty, got: %#v", encoded)
	}
}

func TestNextPageReturnsNilWhenOnLastPage(t *testing.T) {
	current := PaginationResponseMeta{
		Offset: 0,
		Length: 100,
		Total:  200,
		Found:  100,
	}

	got := current.nextPage(nil)

	if got != nil {
		t.Errorf("expected pagination meta for next page to be nil when at end of results, got: %#v", got)
	}
}
func TestCalculatingCursorPositionForNextPage(t *testing.T) {
	current := PaginationResponseMeta{
		Offset: 0,
		Length: 100,
		Total:  200,
		Found:  200,
	}

	expected := PaginationMeta{
		Offset: 100,
		Length: 100,
	}

	got := current.nextPage(nil)

	if got == nil {
		t.Fatalf("did not expect meta for next page to be nil: %#v", got)
	}

	if *got != expected {
		t.Errorf("did not get expected pagination meta: %#v", *got)
	}
}
