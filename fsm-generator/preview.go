package fsm_generator

import (
	"path/filepath"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

func (m *Model) MakePreview() error {
	g := graphviz.New()

	graph, err := g.Graph()
	if err != nil {
		return err
	}

	graph.
		SetLabelLocation(cgraph.TopLocation).
		SetLabel(m.Title + " (" + m.Name + " v." + m.ETag + ")").
		SetPad(0.5).
		SetRankSeparator(2).
		SetNodeSeparator(1)

	prefix := m.Prefix()

	for _, state := range m.States {
		var node *cgraph.Node

		node, err = graph.CreateNode(state.Name)
		if err != nil {
			return err
		}

		node.SetLabel(prefix + state.Name).SetFontSize(28) // + "\n" + state.Description
		node.SetShape(cgraph.RectShape).SetPenWidth(2.5).SetMargin(1)

		if state.Initial {
			node.SetRoot(true)
		}

		if state.SuccessFinal {
			node.SetColor("#66cc00").SetFontColor("#66cc00")
		}

		if state.FailFinal {
			node.SetColor("#ff937a").SetFontColor("#ff937a")
		}

		state.GraphNode = node
	}

	for _, state := range m.States {
		for _, transition := range state.Transitions {
			A := state
			B := transition.State

			var e *cgraph.Edge

			// todo: нумеровать все узлы (1., 2.) и ребра (1.1., 2.5), чтобы было удобно ссылаться в доке
			e, err = graph.CreateEdge(A.Name+"->"+B.Name, A.GraphNode, B.GraphNode)
			if err != nil {
				return err
			}

			e.SetLabel(transition.Condition).
				SetLabelPosition(5, 0).
				SetColor("#888888").
				SetArrowSize(2).
				SetPenWidth(6)

			if B.FailFinal {
				e.SetColor("#dedede")
			}
		}
	}

	err = g.RenderFilename(graph, graphviz.PNG, filepath.Clean(filepath.Join(previewsDir, m.Name+".png")))
	if err != nil {
		return err
	}

	return nil
}
