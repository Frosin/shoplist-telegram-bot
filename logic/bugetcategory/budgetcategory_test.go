package bugetcategory

import (
	"testing"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/bugetstorage"
	"github.com/stretchr/testify/require"
)

func TestCheckSpend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cat  bugetstorage.Category
		now  time.Time
		exp  string
	}{
		{
			name: "over",
			cat: bugetstorage.Category{
				Current: 290,
				Target:  310,
			},
			now: time.Date(2024, 10, 25, 10, 10, 10, 1, time.UTC),
			exp: "🤬 Тормозни! Перерасход на 4 дня",
		},
		{
			name: "less",
			cat: bugetstorage.Category{
				Current: 200,
				Target:  310,
			},
			now: time.Date(2024, 10, 25, 10, 10, 10, 1, time.UTC),
			exp: "",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			res := checkSpend(test.cat, test.now)
			require.Equal(t, test.exp, res)
		})
	}
}
