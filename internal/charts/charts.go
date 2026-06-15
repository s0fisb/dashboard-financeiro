package charts

import (
	"fmt"
	"html/template"
	"math"
	"strings"

	"financial-dashboard/internal/models"
)

// category colors matching dashboard.css palette
var catColors = map[string]string{
	"alimentação": "#f87171",
	"transporte":  "#60a5fa",
	"saúde":       "#34d399",
	"lazer":       "#a78bfa",
	"educação":    "#fbbf24",
	"moradia":     "#f5a623",
	"outros":      "#94a3b8",
}

// DonutSlice is one segment of the donut chart.
type DonutSlice struct {
	Label string
	Value float64
	Color string
}

// RenderDonut builds an inline SVG donut chart (200×200 viewBox).
// The caller embeds the total value separately (in the .donut-center div).
func RenderDonut(byCategory map[string]float64) template.HTML {
	const (
		cx, cy = 100.0, 100.0
		r      = 60.0
	)
	C := 2 * math.Pi * r

	var slices []DonutSlice
	total := 0.0
	for cat, v := range byCategory {
		if v <= 0 {
			continue
		}
		color, ok := catColors[cat]
		if !ok {
			color = "#6b7280"
		}
		slices = append(slices, DonutSlice{Label: cat, Value: v, Color: color})
		total += v
	}

	var sb strings.Builder
	sb.WriteString(`<svg viewBox="0 0 200 200" xmlns="http://www.w3.org/2000/svg" width="100%" height="100%">`)
	// Background track
	sb.WriteString(fmt.Sprintf(
		`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="none" stroke="#1e1e25" stroke-width="22"/>`,
		cx, cy, r,
	))

	if total == 0 {
		sb.WriteString(`</svg>`)
		return template.HTML(sb.String())
	}

	offset := 0.0
	for _, s := range slices {
		arc := (s.Value / total) * C
		dashOffset := -(offset) // rotate so arcs follow each other starting from top
		sb.WriteString(fmt.Sprintf(
			`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="none" stroke="%s" stroke-width="22" `+
				`stroke-dasharray="%.3f %.3f" stroke-dashoffset="%.3f" `+
				`transform="rotate(-90 %.1f %.1f)"/>`,
			cx, cy, r, s.Color,
			arc, C-arc, dashOffset,
			cx, cy,
		))
		offset += arc
	}
	sb.WriteString(`</svg>`)
	return template.HTML(sb.String())
}

func abbrevBRL(v float64) string {
	if v >= 1000 {
		return fmt.Sprintf("R$%.0fk", v/1000)
	}
	return fmt.Sprintf("R$%.0f", v)
}

// RenderBar builds a 580×220 SVG bar chart for monthly trend.
// forecasts[0..n-2] are actuals; forecasts[n-1] is the predicted month.
func RenderBar(forecasts []models.Forecast) template.HTML {
	const (
		svgW, svgH     = 580, 220
		padL, padB     = 55.0, 30.0
		padT, padR     = 15.0, 15.0
		chartW, chartH = svgW - padL - padR, svgH - padB - padT
	)

	n := len(forecasts)
	if n == 0 {
		return template.HTML(`<svg viewBox="0 0 580 220" xmlns="http://www.w3.org/2000/svg"></svg>`)
	}

	// Find max value for scale
	maxVal := 1.0
	for _, f := range forecasts {
		if f.Actual > maxVal {
			maxVal = f.Actual
		}
		if f.Predicted > maxVal {
			maxVal = f.Predicted
		}
		if f.Income > maxVal {
			maxVal = f.Income
		}
	}
	maxVal *= 1.1 // 10% headroom

	groupW := chartW / float64(n)
	barW := groupW * 0.5

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		`<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg" width="100%%" height="100%%">`,
		svgW, svgH,
	))

	// Y-axis guide lines
	for i := 0; i <= 4; i++ {
		pct := float64(i) / 4.0
		y := padT + chartH*(1-pct)
		label := abbrevBRL(maxVal / 1.1 * pct)
		sb.WriteString(fmt.Sprintf(
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#1e1e25" stroke-width="1"/>`,
			padL, y, float64(svgW)-padR, y,
		))
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#5a5450" font-size="9" font-family="monospace" text-anchor="end">%s</text>`,
			padL-4, y+3, label,
		))
	}

	// Income polyline points
	var incPoints []string
	for i, f := range forecasts {
		cx := padL + float64(i)*groupW + groupW/2
		incomeH := (f.Income / maxVal) * chartH
		cy := padT + chartH - incomeH
		incPoints = append(incPoints, fmt.Sprintf("%.1f,%.1f", cx, cy))
	}
	if len(incPoints) > 0 {
		sb.WriteString(fmt.Sprintf(
			`<polyline points="%s" fill="none" stroke="#34d399" stroke-width="2" stroke-dasharray="5,3" opacity="0.7"/>`,
			strings.Join(incPoints, " "),
		))
	}

	// Bars and X-axis labels
	for i, f := range forecasts {
		cx := padL + float64(i)*groupW + groupW/2
		barX := cx - barW/2

		val := f.Actual
		color := "#f87171"
		if i == n-1 && f.Predicted > 0 {
			val = f.Predicted
			color = "#f5a623"
		}
		barH := (val / maxVal) * chartH
		if barH < 1 {
			barH = 1
		}
		barY := padT + chartH - barH

		sb.WriteString(fmt.Sprintf(
			`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="3" fill="%s" opacity="0.85"/>`,
			barX, barY, barW, barH, color,
		))

		// X-axis label
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#5a5450" font-size="9" font-family="monospace" text-anchor="middle">%s</text>`,
			cx, float64(svgH)-padB+12, f.Month,
		))
	}

	// Legend
	sb.WriteString(`<rect x="55" y="5" width="10" height="8" rx="2" fill="#f87171" opacity="0.85"/>`)
	sb.WriteString(`<text x="68" y="12" fill="#a09590" font-size="9" font-family="monospace">Gastos</text>`)
	sb.WriteString(`<rect x="115" y="5" width="10" height="8" rx="2" fill="#f5a623" opacity="0.85"/>`)
	sb.WriteString(`<text x="128" y="12" fill="#a09590" font-size="9" font-family="monospace">Previsão</text>`)
	sb.WriteString(`<line x1="178" y1="9" x2="188" y2="9" stroke="#34d399" stroke-width="2" stroke-dasharray="3,2"/>`)
	sb.WriteString(`<text x="192" y="12" fill="#a09590" font-size="9" font-family="monospace">Renda</text>`)

	sb.WriteString(`</svg>`)
	return template.HTML(sb.String())
}

// RenderForecastLine builds a 580×220 SVG line chart for the forecast tab.
// Shows three series: actual (jade), predicted (gold dashed), income (sky).
func RenderForecastLine(forecasts []models.Forecast) template.HTML {
	const (
		svgW, svgH     = 580, 220
		padL, padB     = 55.0, 30.0
		padT, padR     = 15.0, 15.0
		chartW, chartH = svgW - padL - padR, svgH - padB - padT
	)

	n := len(forecasts)
	if n < 2 {
		return template.HTML(`<svg viewBox="0 0 580 220" xmlns="http://www.w3.org/2000/svg"></svg>`)
	}

	maxVal := 1.0
	for _, f := range forecasts {
		if f.Actual > maxVal {
			maxVal = f.Actual
		}
		if f.Predicted > maxVal {
			maxVal = f.Predicted
		}
		if f.Income > maxVal {
			maxVal = f.Income
		}
	}
	maxVal *= 1.1

	xStep := chartW / float64(n-1)

	point := func(i int, val float64) (float64, float64) {
		x := padL + float64(i)*xStep
		y := padT + chartH*(1-val/maxVal)
		return x, y
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		`<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg" width="100%%" height="100%%">`,
		svgW, svgH,
	))

	// Y-axis guides
	for i := 0; i <= 4; i++ {
		pct := float64(i) / 4.0
		y := padT + chartH*(1-pct)
		sb.WriteString(fmt.Sprintf(
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#1e1e25" stroke-width="1"/>`,
			padL, y, float64(svgW)-padR, y,
		))
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#5a5450" font-size="9" font-family="monospace" text-anchor="end">%s</text>`,
			padL-4, y+3, abbrevBRL(maxVal/1.1*pct),
		))
	}

	// Build point strings for each series
	var actualPts, predictedPts, incomePts []string
	for i, f := range forecasts {
		if f.Actual > 0 {
			x, y := point(i, f.Actual)
			actualPts = append(actualPts, fmt.Sprintf("%.1f,%.1f", x, y))
		}
		if f.Predicted > 0 {
			x, y := point(i, f.Predicted)
			predictedPts = append(predictedPts, fmt.Sprintf("%.1f,%.1f", x, y))
		}
		if f.Income > 0 {
			x, y := point(i, f.Income)
			incomePts = append(incomePts, fmt.Sprintf("%.1f,%.1f", x, y))
		}
	}

	// Income line
	if len(incomePts) > 1 {
		sb.WriteString(fmt.Sprintf(
			`<polyline points="%s" fill="none" stroke="#60a5fa" stroke-width="2" stroke-dasharray="6,3" opacity="0.7"/>`,
			strings.Join(incomePts, " "),
		))
	}
	// Actual line
	if len(actualPts) > 1 {
		sb.WriteString(fmt.Sprintf(
			`<polyline points="%s" fill="none" stroke="#34d399" stroke-width="2"/>`,
			strings.Join(actualPts, " "),
		))
	}
	// Predicted line (dashed gold, typically just one point connected to last actual)
	if len(predictedPts) >= 1 {
		// Connect last actual point to predicted
		var allPts []string
		for i, f := range forecasts {
			if f.Actual > 0 && i == n-2 {
				x, y := point(i, f.Actual)
				allPts = append(allPts, fmt.Sprintf("%.1f,%.1f", x, y))
			}
		}
		allPts = append(allPts, predictedPts...)
		if len(allPts) > 1 {
			sb.WriteString(fmt.Sprintf(
				`<polyline points="%s" fill="none" stroke="#f5a623" stroke-width="2" stroke-dasharray="6,3"/>`,
				strings.Join(allPts, " "),
			))
		}
		// Star marker for prediction point
		for _, p := range predictedPts {
			var px, py float64
			fmt.Sscanf(p, "%f,%f", &px, &py)
			sb.WriteString(fmt.Sprintf(
				`<circle cx="%.1f" cy="%.1f" r="4" fill="#f5a623"/>`,
				px, py,
			))
		}
	}
	// Dots on actual line
	for i, f := range forecasts {
		if f.Actual > 0 {
			x, y := point(i, f.Actual)
			sb.WriteString(fmt.Sprintf(
				`<circle cx="%.1f" cy="%.1f" r="3" fill="#34d399"/>`,
				x, y,
			))
		}
	}

	// X-axis labels
	for i, f := range forecasts {
		x, _ := point(i, 0)
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#5a5450" font-size="9" font-family="monospace" text-anchor="middle">%s</text>`,
			x, float64(svgH)-padB+12, f.Month,
		))
	}

	// Legend
	sb.WriteString(`<line x1="55" y1="9" x2="65" y2="9" stroke="#34d399" stroke-width="2"/>`)
	sb.WriteString(`<text x="68" y="12" fill="#a09590" font-size="9" font-family="monospace">Realizado</text>`)
	sb.WriteString(`<line x1="125" y1="9" x2="135" y2="9" stroke="#f5a623" stroke-width="2" stroke-dasharray="4,2"/>`)
	sb.WriteString(`<text x="138" y="12" fill="#a09590" font-size="9" font-family="monospace">Previsão Lua</text>`)
	sb.WriteString(`<line x1="213" y1="9" x2="223" y2="9" stroke="#60a5fa" stroke-width="2" stroke-dasharray="4,2"/>`)
	sb.WriteString(`<text x="226" y="12" fill="#a09590" font-size="9" font-family="monospace">Renda</text>`)

	sb.WriteString(`</svg>`)
	return template.HTML(sb.String())
}
