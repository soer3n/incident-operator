package cli

import (
	"encoding/json"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	v1 "k8s.io/api/core/v1"
)

const (
	batchSize  = 80         // The number of rows loaded per batch.
	finderPage = "*finder*" // The name of the Finder page.
)

var (
	app         *tview.Application // The tview application.
	pages       *tview.Pages       // The application pages.
	finderFocus tview.Primitive    // The primitive in the Finder that last had focus.
)

func (cli *CLI) RenderPrepareView() (*tview.Application, error) {

	app = tview.NewApplication()

	nodes := tview.NewList().ShowSecondaryText(false)
	nodes.SetBorder(true).SetTitle("Nodes")

	menu := tview.NewTable().SetBorders(true)
	menu.SetBorder(true).SetTitle("Menu")

	workloads := tview.NewList()
	workloads.ShowSecondaryText(false).
		SetDoneFunc(func() {
			workloads.Clear()
			menu.Clear()
			app.SetFocus(nodes)
		})
	workloads.SetBorder(true).SetTitle("Workloads")

	flex := tview.NewFlex().
		AddItem(nodes, 0, 1, true).
		AddItem(workloads, 0, 1, false).
		AddItem(menu, 0, 3, false)

	pages = tview.NewPages().
		AddPage(finderPage, flex, true, true)
	app.SetRoot(pages, true)

	nodeObjs := GetNodes(cli.logger)

	for _, node := range nodeObjs.Items {
		nodes.AddItem(node.Name, "", 0, func() {
			// A database was selected. Show all of its tables.
			workloads.Clear()
			menu.Clear()

			pods := GetPodsByNode("", node.Name, cli.logger)

			for _, pod := range pods.Items {
				workloads.AddItem(pod.Name, "", 0, nil)
			}

			app.SetFocus(workloads)
			workloads.SetChangedFunc(func(i int, tableName string, t string, s rune) {
				menu.Clear()
				menu.SetCell(0, 0, &tview.TableCell{Text: "Name", Align: tview.AlignCenter, Color: tcell.ColorYellow}).
					SetCell(0, 1, &tview.TableCell{Text: "Namespace", Align: tview.AlignCenter, Color: tcell.ColorYellow}).
					SetCell(0, 2, &tview.TableCell{Text: "Annotations", Align: tview.AlignCenter, Color: tcell.ColorYellow}).
					SetCell(0, 3, &tview.TableCell{Text: "Labels", Align: tview.AlignCenter, Color: tcell.ColorYellow}).
					SetCell(0, 4, &tview.TableCell{Text: "HostNetwork", Align: tview.AlignCenter, Color: tcell.ColorYellow})
			})

			color := tcell.ColorWhite

			for ix, pod := range pods.Items {

				rawAnnotations, _ := json.Marshal(pod.GetAnnotations())
				rawLabels, _ := json.Marshal(pod.GetLabels())
				hostNetwork := fmt.Sprintf("%v", pod.Spec.HostNetwork)

				menu.SetCell(ix, 0, &tview.TableCell{Text: pod.Name, Color: color}).
					SetCell(ix+1, 1, &tview.TableCell{Text: pod.GetNamespace(), Color: color}).
					SetCell(ix+1, 2, &tview.TableCell{Text: string(rawAnnotations), Align: tview.AlignRight, Color: color}).
					SetCell(ix+1, 3, &tview.TableCell{Text: string(rawLabels), Align: tview.AlignRight, Color: color}).
					SetCell(ix+1, 4, &tview.TableCell{Text: string(hostNetwork), Align: tview.AlignLeft, Color: color})
			}

			workloads.SetCurrentItem(0)
			workloads.SetChangedFunc(func(i int, podName string, t string, s rune) {
				setContent(node, pods, podName)
			})
		})
	}

	return app, nil
}

func setContent(node v1.Node, pods *v1.PodList, podName string) {

	finderFocus = app.GetFocus()

	if pages.HasPage(node.Name + "." + podName) {
		pages.SwitchToPage(node.Name + "." + podName)
		return
	}

	table := tview.NewTable().
		SetFixed(1, 0).
		SetSeparator(tview.BoxDrawingsLightHorizontal).
		SetBordersColor(tcell.ColorYellow)
	frame := tview.NewFrame(table).
		SetBorders(0, 0, 0, 0, 0, 0)
	frame.SetBorder(true).
		SetTitle(fmt.Sprintf(`Contents of table "%s"`, podName))

	loadRows := func(offset int) {

		table.SetCell(0, 0, &tview.TableCell{Text: "Name", Align: tview.AlignCenter, Color: tcell.ColorYellow}).
			SetCell(0, 1, &tview.TableCell{Text: "Namespace", Align: tview.AlignCenter, Color: tcell.ColorYellow}).
			SetCell(0, 2, &tview.TableCell{Text: "Annotations", Align: tview.AlignCenter, Color: tcell.ColorYellow}).
			SetCell(0, 3, &tview.TableCell{Text: "Labels", Align: tview.AlignCenter, Color: tcell.ColorYellow}).
			SetCell(0, 4, &tview.TableCell{Text: "HostNetwork", Align: tview.AlignCenter, Color: tcell.ColorYellow})

		for ix, pod := range pods.Items {

			rawAnnotations, _ := json.Marshal(pod.GetAnnotations())
			rawLabels, _ := json.Marshal(pod.GetLabels())
			hostNetwork := fmt.Sprintf("%v", pod.Spec.HostNetwork)

			table.SetCell(ix, 0, &tview.TableCell{Text: pod.Name, Color: tcell.ColorDarkCyan}).
				SetCell(ix+1, 1, &tview.TableCell{Text: pod.GetNamespace(), Color: tcell.ColorDarkCyan}).
				SetCell(ix+1, 2, &tview.TableCell{Text: string(rawAnnotations), Align: tview.AlignRight, Color: tcell.ColorDarkCyan}).
				SetCell(ix+1, 3, &tview.TableCell{Text: string(rawLabels), Align: tview.AlignRight, Color: tcell.ColorDarkCyan}).
				SetCell(ix+1, 4, &tview.TableCell{Text: string(hostNetwork), Align: tview.AlignLeft, Color: tcell.ColorDarkCyan})
		}

		frame.Clear()
	}

	loadRows(0)

	table.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEscape:
			// Go back to Finder.
			pages.SwitchToPage(finderPage)
			if finderFocus != nil {
				app.SetFocus(finderFocus)
			}
		case tcell.KeyEnter:
			// Load the next batch of rows.
			loadRows(1)
			table.ScrollToEnd()
		}
	})

	// Add a new page and show it.
	pages.AddPage(node.Name+"."+podName, frame, true, true)
}
