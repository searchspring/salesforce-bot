package metabase

import (
	"fmt"
	"log"
	"testing"

	"github.com/grokify/go-metabase/metabase"
	"github.com/stretchr/testify/require"
)

func createQueryResults() *metabase.DatasetQueryResultsData {
	qr := &metabase.DatasetQueryResultsData{}
	return qr
}

func TestStructFromResult(t *testing.T) {
	mbdao := &DAOImpl{}
	result, err := mbdao.StructFromResult(createQueryResults())
	if err != nil {
		log.Println(err)
		t.Fail()
	}

	require.Contains(t, fmt.Sprint(result.MRR), "-1")
	require.Contains(t, fmt.Sprint(result.FamilyMRR), "-1")
	require.Contains(t, result.Manager, "No company found")
}