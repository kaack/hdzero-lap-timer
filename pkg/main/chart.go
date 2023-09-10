package main

import (
	"bytes"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"gocv.io/x/gocv"
)

func ChartImg(cBuffer *ChartBuffer, mat *gocv.Mat) {

	xValues, yValues := cBuffer.Data()

	graph := chart.Chart{
		Title:  "Activation Area",
		Width:  480,
		Height: 360,
		DPI:    150,
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: xValues,
				YValues: yValues,
			},
		},
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
			Value: 10000,
			Label: "10K",
		}, {
			Value: 20000,
			Label: "20K",
		}, {
			Value: 30000,
			Label: "30K",
		}, {
			Value: 40000,
			Label: "40K",
		}, {
			Value: 50000,
			Label: "50K",
		}},
	}

	buffer := bytes.NewBuffer([]byte{})
	graph.Render(chart.PNG, buffer)

	gocv.IMDecodeIntoMat(buffer.Bytes(), gocv.IMReadAnyDepth, mat)
}

type Point struct {
	x float64
	y float64
}

func graph(dataChan chan Point, matChan chan *gocv.Mat, gMat *gocv.Mat) {
	chartData := NewChartBuffer(100)
	for {
		select {
		case point := <-dataChan:
			chartData.Push(point.x, point.y)
			ChartImg(chartData, gMat)
			matChan <- gMat
		}
	}
}
