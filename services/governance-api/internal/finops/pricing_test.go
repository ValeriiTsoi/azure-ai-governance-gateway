package finops

import "testing"

func TestStaticCatalogLookupNormalizesProviderAndModel(
	t *testing.T,
) {
	catalog, err := NewStaticCatalog(
		[]Rate{
			{
				Provider:            "azure-openai",
				Model:               "test-model",
				InputPerMillionUSD:  1,
				OutputPerMillionUSD: 2,
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	rate, found := catalog.Lookup(
		" Azure-OpenAI ",
		" TEST-MODEL ",
	)

	if !found {
		t.Fatal("expected pricing rate")
	}

	if rate.InputPerMillionUSD != 1 {
		t.Fatalf(
			"unexpected input price: %f",
			rate.InputPerMillionUSD,
		)
	}

	if rate.OutputPerMillionUSD != 2 {
		t.Fatalf(
			"unexpected output price: %f",
			rate.OutputPerMillionUSD,
		)
	}
}

func TestStaticCatalogRejectsDuplicateRate(
	t *testing.T,
) {
	_, err := NewStaticCatalog(
		[]Rate{
			{
				Provider: "azure-openai",
				Model:    "test-model",
			},
			{
				Provider: " AZURE-OPENAI ",
				Model:    " TEST-MODEL ",
			},
		},
	)

	if err == nil {
		t.Fatal("expected duplicate pricing error")
	}
}

func TestStaticCatalogRejectsNegativePrice(
	t *testing.T,
) {
	_, err := NewStaticCatalog(
		[]Rate{
			{
				Provider:            "test-provider",
				Model:               "test-model",
				InputPerMillionUSD:  -1,
				OutputPerMillionUSD: 2,
			},
		},
	)

	if err == nil {
		t.Fatal("expected negative pricing error")
	}
}
