package main

import (
	"bytes"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"gocv.io/x/gocv"
)

func ChartImg(buffers []*ChartBuffer, mat *gocv.Mat) {

	graph := chart.Chart{
		Title:  "Activation Area",
		Width:  480,
		Height: 360,
		DPI:    150,
	}

	for _, buffer := range buffers {
		xValues, yValues := buffer.Data()
		series := chart.ContinuousSeries{
			XValues: xValues,
			YValues: yValues,
		}

		graph.Series = append(graph.Series, series)
	}

	graph.Background = chart.Style{
		Padding: chart.Box{
			Top:    25,
			Left:   25,
			Right:  25,
			Bottom: 25,
		},
		FillColor: drawing.ColorFromHex("efefef"),
	}

	graph.YAxis = chart.YAxis{
		Style: chart.Style{
			Show: true,
		},
		Ticks: []chart.Tick{{
			Value: 0,
			Label: "0",
		}, {
			Value: 100000,
			Label: "100K",
		}},
	}

	buffer := bytes.NewBuffer([]byte{})
	_ = graph.Render(chart.PNG, buffer)
	_ = gocv.IMDecodeIntoMat(buffer.Bytes(), gocv.IMReadAnyDepth, mat)
}

type Point struct {
	x float64
	y float64
}

func graph(dataChan chan []Point, matChan chan *gocv.Mat, gMat *gocv.Mat, gates []*Gate) {

	var chartsData []*ChartBuffer
	for range gates {
		chartsData = append(chartsData, NewChartBuffer(100))
	}

	for {
		select {
		case points := <-dataChan:
			for i := range gates {
				chartsData[i].Push(points[i].x, points[i].y)
			}
			ChartImg(chartsData, gMat)
			matChan <- gMat
		}
	}
}
