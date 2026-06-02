package telegram

import (
	"strings"
)

type profileParams map[string]interface{}

func (t *Telegram) Profile(profileParams profileParams) string {
	what := strings.Builder{}
	what.WriteString("<b>Профиль</b>")
	what.WriteString("<blockquote>")

	if val, ok := profileParams["id"]; ok {
		what.WriteString("ID: " + val.(string))
	}

	if val, ok := profileParams["on_work_text"]; ok {
		what.WriteString("\nСмена: " + val.(string))
	}

	if val, ok := profileParams["fullname"]; ok {
		what.WriteString("\nИмя: " + val.(string))
	}
	if val, ok := profileParams["phone"]; ok {
		what.WriteString("\nНомер телефона: " + val.(string))
	}

	if val, ok := profileParams["verified"]; ok {
		what.WriteString("\nВерифицирован: " + val.(string))
	}
	if val, ok := profileParams["rating"]; ok {
		what.WriteString("\nРейтинг: " + val.(string))
	}
	what.WriteString("</blockquote>")

	if val, ok := profileParams["has_active_order"]; ok {
		if val.(bool) {
			what.WriteString("\n\n<b>Активный заказ</b>")
			what.WriteString("<blockquote>")
			if val, ok := profileParams["from_address"]; ok {
				what.WriteString("От: " + val.(string))
			}
			if val, ok := profileParams["to_address"]; ok {
				what.WriteString("\nДо: " + val.(string))
			}
			what.WriteString("</blockquote>")
		}
	}

	return what.String()
}
