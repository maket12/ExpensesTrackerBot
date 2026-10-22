package mapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatTelegramHTML(t *testing.T) {
	t.Run("converts commonmark double asterisks to bold tag", func(t *testing.T) {
		input := "За этот месяц вы потратили **320 ₽** (кофе 220 ₽, такси 100 ₽)."
		expected := "За этот месяц вы потратили <b>320 ₽</b> (кофе 220 ₽, такси 100 ₽)."
		assert.Equal(t, expected, FormatTelegramHTML(input))
	})

	t.Run("converts double underscores to bold tag", func(t *testing.T) {
		input := "Итого: __1500 ₽__"
		expected := "Итого: <b>1500 ₽</b>"
		assert.Equal(t, expected, FormatTelegramHTML(input))
	})

	t.Run("converts multiple bold segments", func(t *testing.T) {
		input := "Потрачено: **100 ₽** на еду и **200 ₽** на такси"
		expected := "Потрачено: <b>100 ₽</b> на еду и <b>200 ₽</b> на такси"
		assert.Equal(t, expected, FormatTelegramHTML(input))
	})

	t.Run("converts inline code", func(t *testing.T) {
		input := "Расход с ID `b9e07fb8-3333` удалён"
		expected := "Расход с ID <code>b9e07fb8-3333</code> удалён"
		assert.Equal(t, expected, FormatTelegramHTML(input))
	})

	t.Run("converts code block", func(t *testing.T) {
		input := "```\nexpense id\n```"
		expected := "<pre>expense id</pre>"
		assert.Equal(t, expected, FormatTelegramHTML(input))
	})

	t.Run("converts bullet list asterisks to dots", func(t *testing.T) {
		input := "* кофе 100 ₽\n* чай 50 ₽"
		expected := "• кофе 100 ₽\n• чай 50 ₽"
		assert.Equal(t, expected, FormatTelegramHTML(input))
	})

	t.Run("escapes raw html brackets in user content", func(t *testing.T) {
		input := "Траты: 100 < 200 & 50 > 20"
		expected := "Траты: 100 &lt; 200 &amp; 50 &gt; 20"
		assert.Equal(t, expected, FormatTelegramHTML(input))
	})

	t.Run("preserves existing telegram html tags", func(t *testing.T) {
		input := "<b>Заголовок</b>: <i>курсив</i> и <code>код</code>"
		expected := "<b>Заголовок</b>: <i>курсив</i> и <code>код</code>"
		assert.Equal(t, expected, FormatTelegramHTML(input))
	})

	t.Run("handles empty string", func(t *testing.T) {
		assert.Equal(t, "", FormatTelegramHTML(""))
		assert.Equal(t, "   ", FormatTelegramHTML("   "))
	})
}
