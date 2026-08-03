package site

import (
	"math"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/araihu/goshtoso-charts/components/interactive"
)

const heartLine3DFormula = "x = 16 × sin³(t); y = 2.4 × sin(3t); z = 13 × cos(t) − 5 × cos(2t) − 2 × cos(3t) − cos(4t)"

// PajeWorkflowGraph renders Pajé's durable five-stage code workflow.
func PajeWorkflowGraph(project Project) interactive.Instance {
	return interactive.Graph(interactive.GraphConfig{
		Label:  project.Name + " — " + project.Description,
		Layout: interactive.GraphLayoutNone,
		Nodes: []interactive.Node{
			graphNode("resolve", 14, 35, 42, "#ff8a3d"),
			graphNode("execute", 34, 18, 54, "#ff8a3d"),
			graphNode("approval", 56, 50, 48, "#ef476f"),
			graphNode("publish", 79, 28, 46, "#c7ff4a"),
			graphNode("finalize", 24, 74, 38, "#ff8a3d"),
		},
		Links: []interactive.Link{
			{Source: "resolve", Target: "execute"},
			{Source: "execute", Target: "approval"},
			{Source: "approval", Target: "publish"},
			{Source: "resolve", Target: "finalize"},
			{Source: "finalize", Target: "approval"},
		},
		Draggable: interactive.Bool(false),
		Width:     "100%",
		Height:    "100%",
		Options: interactive.ChartOptions{
			Animation: interactive.Bool(false),
			Legend:    &interactive.LegendOptions{Show: interactive.Bool(false)},
			Tooltip:   &interactive.TooltipOptions{Show: interactive.Bool(false)},
			Controls:  chartcontrol.Options{Expand: chartcontrol.Bool(false)},
			Export:    &chartcontrol.ExportOptions{Disabled: true},
		},
		SeriesOptions: interactive.SeriesOptions{
			Label:     &interactive.LabelOptions{Show: interactive.Bool(false)},
			LineStyle: &interactive.LineStyle{Color: "#718397", Width: 2},
		},
		Style: charttheme.Style{Palette: charttheme.PaletteAraiHu, Class: "paje-workflow-chart"},
	})
}

// X9AvailabilityChart renders the same stacked availability model used by the
// Goshtoso Charts live-availability example. Client-side ticks keep the static
// site moving without inventing a separate chart implementation.
func X9AvailabilityChart(project Project) interactive.Instance {
	categories, healthy, degraded, down := availabilitySnapshot(0)
	return interactive.Bar(interactive.BarConfig{
		Label: project.Name + " — " + project.Description,
		XAxis: categories,
		Series: []interactive.BarSeries{
			{Name: "Healthy", Data: healthy},
			{Name: "Degraded", Data: degraded},
			{Name: "Down", Data: down},
		},
		Width:  "100%",
		Height: "100%",
		Options: interactive.ChartOptions{
			Animation: interactive.Bool(false),
			Legend:    &interactive.LegendOptions{Show: interactive.Bool(true), Top: "4%"},
			Tooltip:   &interactive.TooltipOptions{Show: interactive.Bool(false)},
			XAxis: &interactive.AxisOptions{
				LabelInterval:  interactive.Int(5),
				ShowFirstLabel: interactive.Bool(true),
				ShowLastLabel:  interactive.Bool(true),
			},
			YAxis: &interactive.AxisOptions{
				Min:  interactive.Float(0),
				Max:  interactive.Float(1),
				Show: interactive.Bool(false),
			},
			Controls: chartcontrol.Options{Expand: chartcontrol.Bool(false)},
			Export:   &chartcontrol.ExportOptions{Disabled: true},
		},
		SeriesOptions: interactive.SeriesOptions{
			Animation: interactive.Bool(false),
			Stack:     "availability",
			BarWidth:  "70%",
			BarGap:    "0%",
		},
		Style: charttheme.Style{
			Palette: charttheme.PaletteStatus,
			Colors:  []string{"#70ad84", "#bd8c43", "#bd6270"},
			Class:   "x9-availability-chart",
		},
	})
}

// GoshtosoHeartLine3D renders the parametric heart as actual Line3D data.
func GoshtosoHeartLine3D(project Project, label string) interactive.Instance {
	return interactive.Line3D(interactive.Line3DConfig{
		Label: project.Name + " — " + label,
		Series: []interactive.Line3DSeries{{
			Name:   "Heart",
			Points: heartLine3DPoints(),
			Color:  "#c7ff4a",
			Options: interactive.SeriesOptions{
				Animation: interactive.Bool(false),
				LineStyle: &interactive.LineStyle{Width: 10, Opacity: interactive.Float(.96)},
			},
		}},
		Grid: interactive.Line3DGrid{
			Width: 115, Height: 100, Depth: 40,
			View: &interactive.Line3DView{AutoRotate: interactive.Bool(true), AutoRotateSpeed: 8},
		},
		DataSummary: interactive.Line3DDataSummary{
			Formula: heartLine3DFormula, Parameter: "t", ParameterMin: 0, ParameterMax: 2 * math.Pi,
		},
		Width:  "100%",
		Height: "100%",
		Options: interactive.ChartOptions{
			Animation: interactive.Bool(false),
			Legend:    &interactive.LegendOptions{Show: interactive.Bool(false)},
			Tooltip:   &interactive.TooltipOptions{Show: interactive.Bool(false)},
			Controls:  chartcontrol.Options{Expand: chartcontrol.Bool(false)},
			Export:    &chartcontrol.ExportOptions{Disabled: true},
		},
		Style: charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#c7ff4a"}, Class: "goshtoso-heart-chart"},
	})
}

func heartLine3DPoints() []interactive.Point3D {
	const segments = 320
	points := make([]interactive.Point3D, segments+1)
	for index := range points {
		t := float64(index) / segments * 2 * math.Pi
		points[index] = interactive.Point3D{
			X: 16 * math.Pow(math.Sin(t), 3),
			Y: 2.4 * math.Sin(3*t),
			Z: 13*math.Cos(t) - 5*math.Cos(2*t) - 2*math.Cos(3*t) - math.Cos(4*t),
		}
	}
	return points
}

func availabilitySnapshot(step int) ([]string, []interactive.BarData, []interactive.BarData, []interactive.BarData) {
	const buckets = 36
	categories := make([]string, buckets)
	healthy := make([]interactive.BarData, buckets)
	degraded := make([]interactive.BarData, buckets)
	down := make([]interactive.BarData, buckets)
	for index := range categories {
		categories[index] = "--:--:--"
		state := availabilityState(step + index)
		healthy[index].Value = availabilityValue(state == 0)
		degraded[index].Value = availabilityValue(state == 1)
		down[index].Value = availabilityValue(state == 2)
	}
	return categories, healthy, degraded, down
}

func availabilityState(value int) int {
	switch phase := value % 24; {
	case phase >= 8 && phase <= 10:
		return 1
	case phase >= 17 && phase <= 19:
		return 2
	default:
		return 0
	}
}

func availabilityValue(active bool) float64 {
	if active {
		return 1
	}
	return 0
}

func graphNode(name string, x, y, size float64, color string) interactive.Node {
	return interactive.Node{
		Name: name,
		X:    &x,
		Y:    &y,
		Size: size,
		ItemStyle: &interactive.ItemStyle{
			Color:       color,
			BorderColor: "#07111f",
			BorderWidth: 2,
		},
	}
}
