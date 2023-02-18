package handlers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMsgFilter(t *testing.T) {
	testData := [][]string{
		[]string{"@万松园的胖子 我想知道你是谁", "我想知道你是谁"},
	}

	for _, item := range testData {
		from := item[0]
		expect := item[1]

		ret := msgFilter(from)
		assert.Equal(t, expect, ret)
	}

}
