package twiml

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullInt(t *testing.T) {
	type testDoc struct {
		XMLName   xml.Name `xml:"Test"`
		HasValue  NullInt  `xml:",attr"`
		HasValue2 NullInt  `xml:",attr"`
		NoValue   NullInt  `xml:",attr"`
	}
	var doc testDoc
	doc.HasValue = NullInt{Valid: true}
	doc.HasValue2 = NullInt{Valid: true, Value: 2}

	data, err := xml.Marshal(&doc)
	assert.NoError(t, err)
	assert.Equal(t, `<Test HasValue="0" HasValue2="2"></Test>`, string(data))

	var doc2 testDoc
	err = xml.Unmarshal(data, &doc2)
	assert.NoError(t, err)
	assert.True(t, doc2.HasValue.Valid)
	assert.NotNil(t, doc2.HasValue2)
	assert.False(t, doc2.NoValue.Valid)
	assert.Equal(t, 2, doc2.HasValue2.Value)
}
