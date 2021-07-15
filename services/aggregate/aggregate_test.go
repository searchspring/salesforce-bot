package aggregate

import (
	"testing"

	"github.com/searchspring/nebo/models"
	"github.com/stretchr/testify/require"
)

func metabaseCustomers() []*models.AccountInfo {
	return []*models.AccountInfo{
		{
			SiteId:  "123456",
			Website: "one.com",
		},
		{
			SiteId:  "abcdef",
			Website: "two.com",
		},
	}
}

func salesforceCustomers() []*models.AccountInfo {
	return []*models.AccountInfo{
		{
			Type:    "Customer",
			SiteId:  "123abc",
			Website: "three.com",
		},
		{
			Type:    "Prospect",
			SiteId:  "123456",
			Website: "one.com",
		},
	}
}

func TestAddingMetabaseAccounts(t *testing.T) {
	metabaseAccounts := addMetabaseAccounts(metabaseCustomers(), salesforceCustomers())

	require.Equal(t, 1, len(metabaseAccounts))
	require.Equal(t, "abcdef", metabaseAccounts[0].SiteId)
	require.Equal(t, "two.com", metabaseAccounts[0].Website)
}

func TestAddingSalesforceAccounts(t *testing.T) {
	combinedAccounts := addSalesforceAccounts(metabaseCustomers(), salesforceCustomers())

	require.Equal(t, 3, len(combinedAccounts))
	require.Equal(t, "123456", combinedAccounts[0].SiteId)
	require.Equal(t, "one.com", combinedAccounts[0].Website)
	require.Equal(t, "abcdef", combinedAccounts[1].SiteId)
	require.Equal(t, "two.com", combinedAccounts[1].Website)
	require.Equal(t, "123abc", combinedAccounts[2].SiteId)
	require.Equal(t, "three.com", combinedAccounts[2].Website)
}

func TestAddingSFAndMBAccounts(t *testing.T) {
	combinedAccounts := addMetabaseAccounts(metabaseCustomers(), salesforceCustomers())
	combinedAccounts = addSalesforceAccounts(combinedAccounts, salesforceCustomers())

	require.Equal(t, 2, len(combinedAccounts))
	require.Equal(t, "abcdef", combinedAccounts[0].SiteId)
	require.Equal(t, "two.com", combinedAccounts[0].Website)
	require.Equal(t, "123abc", combinedAccounts[1].SiteId)
	require.Equal(t, "three.com", combinedAccounts[1].Website)
}
