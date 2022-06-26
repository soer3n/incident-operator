package cli

import (
	"context"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/soer3n/incident-operator/api/v1alpha1"
	"github.com/soer3n/incident-operator/internal/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

func (cli *CLI) setNodeSettings(name string, hostMenu *tview.Table) {

	hostMenu.SetCell(0, 0, &tview.TableCell{Text: "Name", Align: tview.AlignLeft, Color: tcell.ColorYellow, NotSelectable: true}).
		SetCell(0, 1, &tview.TableCell{Text: "Value", Align: tview.AlignLeft, Color: tcell.ColorYellow, NotSelectable: true}).
		SetCell(0, 2, &tview.TableCell{Text: "Description", Align: tview.AlignLeft, Color: tcell.ColorYellow, NotSelectable: true})

	isolate := "false"
	flags := cli.q.Spec.Flags

	for _, node := range cli.q.Spec.Nodes {
		if node.Name == name {
			isolate = strconv.FormatBool(node.Isolate)

			if node.Isolate {
				flags = node.Flags
			}

			break
		}

	}

	settings := []cellEntry{
		{key: "isolate", value: isolate, desc: ""},
		{key: "debug", value: strconv.FormatBool(cli.q.Spec.Debug.Enabled), desc: ""},
		{key: "disableEviction", value: strconv.FormatBool(*flags.DisableEviction), desc: ""},
		{key: "deleteEmptyDirData", value: strconv.FormatBool(*flags.DeleteEmptyDirData), desc: ""},
		{key: "ignoreAllDaemonSets", value: strconv.FormatBool(*flags.IgnoreAllDaemonSets), desc: ""},
		{key: "force", value: strconv.FormatBool(*flags.Force), desc: ""},
	}

	for ix, iv := range settings {

		hostMenu.SetCell(ix+1, 0, &tview.TableCell{Text: iv.key, Align: tview.AlignLeft, Color: tcell.ColorWhite, NotSelectable: true}).
			SetCell(ix+1, 1, &tview.TableCell{Text: iv.value, Align: tview.AlignLeft, Color: tcell.ColorWhite, NotSelectable: false}).
			SetCell(ix+1, 2, &tview.TableCell{Text: iv.desc, Align: tview.AlignLeft, Color: tcell.ColorWhite, NotSelectable: true})
	}

	hostMenu.SetBorder(true).SetTitle("Settings").SetTitleAlign(tview.AlignCenter)

}

func (cli *CLI) updateNodesSettings(cell *tview.TableCell, value string, nodes *tview.List) {

	ix := nodes.GetCurrentItem()
	nodeSelection, _ := nodes.GetItemText(ix)

	for k, n := range cli.q.Spec.Nodes {
		if n.Name == nodeSelection {
			switch cell.Text {
			case "isolate":
				cli.q.Spec.Nodes[k].Isolate, _ = strconv.ParseBool(value)
			case "debug":
				cli.q.Spec.Debug.Enabled, _ = strconv.ParseBool(value)
			case "disableEviction":
				v, _ := strconv.ParseBool(value)
				cli.q.Spec.Nodes[k].Flags.DisableEviction = pointer.Bool(v)
			case "deleteEmptyDirData":
				v, _ := strconv.ParseBool(value)
				cli.q.Spec.Nodes[k].Flags.DeleteEmptyDirData = pointer.Bool(v)
			case "ignoreAllDaemonSets":
				v, _ := strconv.ParseBool(value)
				cli.q.Spec.Nodes[k].Flags.IgnoreAllDaemonSets = pointer.Bool(v)
			case "force":
				v, _ := strconv.ParseBool(value)
				cli.q.Spec.Nodes[k].Flags.Force = pointer.Bool(v)
			}
		}
	}

}

func (cli *CLI) setWorkloadsSettings(mainText string, menu *tview.Table) {

	pods := cli.getPodsByNode("", mainText)

	menu.SetCell(0, 0, &tview.TableCell{Text: "Name", Align: tview.AlignLeft, Color: tcell.ColorYellow, NotSelectable: true}).
		SetCell(0, 1, &tview.TableCell{Text: "Namespace", Align: tview.AlignLeft, Color: tcell.ColorYellow, NotSelectable: true}).
		SetCell(0, 2, &tview.TableCell{Text: "Controlled by", Align: tview.AlignLeft, Color: tcell.ColorYellow, NotSelectable: true})

	color := tcell.ColorWhite

	for ix, pod := range pods.Items {

		owners := pod.ObjectMeta.GetOwnerReferences()
		ownerTypes := ""
		notSelectable := true

		for _, o := range owners {
			owner := o.Kind
			if ownerTypes != "" {
				ownerTypes = ownerTypes + ","
			}
			if o.Kind == "ReplicaSet" {
				owner = "Deployment"
			}
			ownerTypes = ownerTypes + owner
		}

		if strings.Contains(ownerTypes, "Deployment") || strings.Contains(ownerTypes, "DaemonSet") {
			notSelectable = false
		}

		menu.SetCell(ix+1, 0, &tview.TableCell{Text: pod.Name, Color: color, NotSelectable: notSelectable}).
			SetCell(ix+1, 1, &tview.TableCell{Text: pod.GetNamespace(), Color: color, NotSelectable: notSelectable}).
			SetCell(ix+1, 2, &tview.TableCell{Text: ownerTypes, Align: tview.AlignRight, Color: color, NotSelectable: notSelectable})
	}

}

func (cli *CLI) setWorkloadSettings(workload, namespace, node string, menu, workloads *tview.Table) {

	menu.SetCell(0, 0, &tview.TableCell{Text: "Name", Align: tview.AlignLeft, Color: tcell.ColorYellow, NotSelectable: true}).
		SetCell(0, 1, &tview.TableCell{Text: "Value", Align: tview.AlignLeft, Color: tcell.ColorYellow, NotSelectable: true}).
		SetCell(0, 2, &tview.TableCell{Text: "Description", Align: tview.AlignLeft, Color: tcell.ColorYellow, NotSelectable: true})

	color := tcell.ColorWhite

	r, _ := workloads.GetSelection()
	workloadSelection := workloads.GetCell(r, 0)

	ownerName, ownerType := cli.getOwner(workloadSelection.Text, namespace)

	workloadObj := v1alpha1.Resource{
		Name:      ownerName,
		Namespace: namespace,
		Keep:      false,
		Type:      ownerType,
	}

	for _, iv := range cli.q.Spec.Nodes {
		if iv.Name == node {
			for _, sv := range iv.Resources {
				if sv.Name == workloadObj.Name {
					workloadObj = sv
				}
			}
		}
	}

	settings := []cellEntry{
		{key: "keep", value: strconv.FormatBool(workloadObj.Keep), desc: ""},
	}

	for ix, iv := range settings {

		menu.SetCell(ix+1, 0, &tview.TableCell{Text: iv.key, Color: color, NotSelectable: true}).
			SetCell(ix+1, 1, &tview.TableCell{Text: iv.value, Color: color}).
			SetCell(ix+1, 2, &tview.TableCell{Text: iv.desc, Align: tview.AlignRight, Color: color, NotSelectable: true})
	}
}

func (cli *CLI) getOwner(podName, namespace string) (string, string) {

	var rs *appsv1.ReplicaSet
	var pod *corev1.Pod
	var err error

	c := utils.GetTypedKubernetesClient()

	getOpts := metav1.GetOptions{}

	if pod, err = c.CoreV1().Pods(namespace).Get(context.TODO(), podName, getOpts); err != nil {
		cli.logger.Error(err)
		return "", ""
	}

	ownerName := pod.ObjectMeta.OwnerReferences[0].Name
	ownerType := pod.ObjectMeta.OwnerReferences[0].Kind

	switch pod.ObjectMeta.OwnerReferences[0].Kind {
	case "ReplicaSet":
		if rs, err = c.AppsV1().ReplicaSets(namespace).Get(context.TODO(), pod.ObjectMeta.OwnerReferences[0].Name, getOpts); err != nil {
			cli.logger.Error(err)
		}
		ownerName = rs.ObjectMeta.OwnerReferences[0].Name
		ownerType = rs.ObjectMeta.OwnerReferences[0].Kind
	}

	return ownerName, ownerType
}

func (cli *CLI) updateWorkloadSettings(cell *tview.TableCell, value, namespace string, workloads *tview.Table, nodes *tview.List) {

	r, _ := workloads.GetSelection()
	workloadSelection := workloads.GetCell(r, 0)

	ix := nodes.GetCurrentItem()
	nodeSelection, _ := nodes.GetItemText(ix)

	ownerName, ownerType := cli.getOwner(workloadSelection.Text, namespace)

	list := []v1alpha1.Resource{}
	key := 0

	for k, n := range cli.q.Spec.Nodes {
		if n.Name == nodeSelection {
			list = n.Resources
			key = k
		}
	}

	if len(list) == 0 {
		list = append(list, v1alpha1.Resource{Name: ownerName, Namespace: namespace, Type: ownerType})
	}

	itemIsPresent := false

	for k, w := range list {
		if ownerName == w.Name {
			switch cell.Text {
			case "keep":
				list[k].Keep, _ = strconv.ParseBool(value)
				cli.q.Spec.Nodes[key].Resources = list
				itemIsPresent = true
			}
		}
	}

	if !itemIsPresent {
		keep, _ := strconv.ParseBool(value)
		cli.q.Spec.Nodes[key].Resources = append(cli.q.Spec.Nodes[key].Resources, v1alpha1.Resource{Name: ownerName, Namespace: namespace, Type: ownerType, Keep: keep})
	}

}

func (cli *CLI) initMainMenu(nodes, workloads *tview.List, menu, hostMenu *tview.Table) *tview.List {

	nodes.SetBorder(true).SetTitle("Nodes")

	nodeObjs := cli.getNodes()

	for _, node := range nodeObjs.Items {
		nodes.AddItem(node.Name, "", 0, nil)

	}

	nodes.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		// A node was selected. Show all of its pods.
		ix := workloads.GetCurrentItem()
		workloadSelection, _ := workloads.GetItemText(ix)
		menu.Clear()
		hostMenu.Blur()

		switch workloadSelection {
		case "host":
			cli.setNodeSettings(mainText, hostMenu)
		case "workloads":
			cli.setWorkloadsSettings(mainText, menu)
		}
	})

	nodes.SetSelectedFunc(
		func(index int, mainText string, secondaryText string, shortcut rune) {
			app.SetFocus(workloads)
		})

	nodes.SetCurrentItem(0)

	return nodes
}

func (cli *CLI) initNodeMenu(pages *tview.Pages, nodes, nodeMenu *tview.List, menu, hostMenu *tview.Table) *tview.List {

	nodeMenu.ShowSecondaryText(false).
		SetDoneFunc(func() {
			nodeMenu.Clear()
			menu.Clear()
			hostMenu.Blur()
			app.SetFocus(nodes)
		})

	nodeMenu.SetBorder(true).SetTitle("Menu")

	menuItems := []string{"host", "workloads"}

	for _, item := range menuItems {
		nodeMenu.AddItem(item, "", 0, nil)
	}

	nodeMenu.SetChangedFunc(func(i int, menuItem string, t string, s rune) {
		menu.Clear()
		hostMenu.Blur()
		current, _ := pages.GetFrontPage()

		if current != menuItem {
			pages.SwitchToPage(menuItem)
			app.SetFocus(nodeMenu)
		}

		ix := nodes.GetCurrentItem()
		nodeSelection, _ := nodes.GetItemText(ix)

		switch menuItem {
		case "host":
			cli.setNodeSettings(nodeSelection, hostMenu)
		case "workloads":
			cli.setWorkloadsSettings(nodeSelection, menu)
		}
	})

	nodeMenu.SetSelectedFunc(func(i int, podName string, t string, s rune) {

		ix := nodeMenu.GetCurrentItem()
		nodeMenuSelection, _ := nodeMenu.GetItemText(ix)

		switch nodeMenuSelection {
		case "host":
			app.SetFocus(hostMenu)
		case "workloads":
			app.SetFocus(menu)
			//TODO: set previous selected item
			menu.Select(1, 1).SetFixed(1, 1).SetSelectable(true, false)
		}

	})

	nodeMenu.SetDoneFunc(func() {

		if nodeMenu.HasFocus() {
			app.SetFocus(nodes)
		}
	})

	return nodeMenu
}

func (cli *CLI) initHostSettingsMenu(pages *tview.Pages, hostMenu *tview.Table, nodes, workloads, boxHost *tview.List) *tview.Table {

	hostMenu.SetBorder(true).SetTitle("Settings")

	hostMenu.SetCell(0, 0, &tview.TableCell{Text: "Name", Align: tview.AlignLeft, Color: tcell.ColorYellow, NotSelectable: true}).
		SetCell(0, 1, &tview.TableCell{Text: "Value", Align: tview.AlignLeft, Color: tcell.ColorYellow, NotSelectable: true}).
		SetCell(0, 2, &tview.TableCell{Text: "Description", Align: tview.AlignLeft, Color: tcell.ColorYellow, NotSelectable: true})

	ix := nodes.GetCurrentItem()
	nodeSelection, _ := nodes.GetItemText(ix)

	current := v1alpha1.Node{
		Name:    nodeSelection,
		Isolate: false,
		Flags: v1alpha1.Flags{
			IgnoreAllDaemonSets: pointer.BoolPtr(false),
			DeleteEmptyDirData:  pointer.BoolPtr(false),
			DisableEviction:     pointer.BoolPtr(false),
			IgnoreErrors:        pointer.Bool(false),
			Force:               pointer.BoolPtr(false),
		},
	}

	for _, iv := range cli.q.Spec.Nodes {
		if iv.Name == nodeSelection {
			current = iv
			break
		}
	}

	settings := []cellEntry{
		{key: "isolate", value: strconv.FormatBool(current.Isolate), desc: ""},
		{key: "debug", value: strconv.FormatBool(cli.q.Spec.Debug.Enabled), desc: ""},
		{key: "disableEviction", value: strconv.FormatBool(*cli.q.Spec.Flags.DisableEviction), desc: ""},
		{key: "deleteEmptyDirData", value: strconv.FormatBool(*cli.q.Spec.Flags.DeleteEmptyDirData), desc: ""},
		{key: "ignoreAllDaemonSets", value: strconv.FormatBool(*cli.q.Spec.Flags.IgnoreAllDaemonSets), desc: ""},
		{key: "force", value: strconv.FormatBool(*cli.q.Spec.Flags.Force), desc: ""},
	}

	color := tcell.ColorWhite

	for i, ix := range settings {

		hostMenu.SetCell(i+1, 0, &tview.TableCell{Text: ix.key, Color: color, NotSelectable: true}).
			SetCell(i+1, 1, &tview.TableCell{Text: ix.value, Color: color}).
			SetCell(i+1, 2, &tview.TableCell{Text: ix.desc, Align: tview.AlignRight, Color: color, NotSelectable: true})
	}

	hostMenu.SetBorder(true).SetTitle("Settings").SetTitleAlign(tview.AlignCenter)

	hostMenu.Select(1, 1).SetSelectable(true, true)

	hostMenu.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEscape:
			pages.SwitchToPage(finderHostPage).HidePage("modalHost")
			app.SetFocus(workloads)
		}
	}).SetSelectedFunc(func(row int, column int) {
		//TODO: set selected item for quarantine resource
		pages.ShowPage("modalHost")
		app.SetFocus(boxHost)
	})

	return hostMenu

}

func (cli *CLI) initModalSelectionFunc() func(p tview.Primitive, width, height int) tview.Primitive {

	return func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewGrid().
			SetColumns(0, width, 0).
			SetRows(0, height, 0).
			SetBorders(true).
			AddItem(p, 1, 1, 1, 1, 0, 0, true)
	}

}

func (cli *CLI) initModalSelectionBox(pages *tview.Pages, name, pageToSwitch string, options []string, workloadMenu, workloads *tview.Table, nodes, nodeMenu *tview.List) *tview.List {

	box := tview.NewList().SetHighlightFullLine(true)
	box.SetBorder(true).SetTitle("Options")
	box.SetSelectedFunc(
		func(index int, mainText string, secondaryText string, shortcut rune) {
			ix := nodeMenu.GetCurrentItem()
			nodeMenuSelection, _ := nodeMenu.GetItemText(ix)

			r, c := workloadMenu.GetSelection()
			cell := workloadMenu.GetCell(r, c-1)

			r, _ = workloads.GetSelection()
			nsCell := workloads.GetCell(r, 1)

			ix = nodes.GetCurrentItem()
			nodeSelection, _ := nodes.GetItemText(ix)

			switch nodeMenuSelection {
			case "host":
				cli.updateNodesSettings(cell, mainText, nodes)
				cli.setNodeSettings(nodeSelection, workloadMenu)
			case "workloads":
				cli.updateWorkloadSettings(cell, mainText, nsCell.Text, workloads, nodes)
				cli.setWorkloadSettings(cell.Text, nsCell.Text, nodeSelection, workloadMenu, workloads)
			}

			pages.ShowPage(pageToSwitch).HidePage(name)
			app.SetFocus(workloadMenu)
		})

	for _, item := range options {
		box.AddItem(item, "", 0, nil)

	}

	return box

}

func (cli *CLI) initWorkloadsMenu(pages *tview.Pages, menu *tview.Table, nodes, workloads *tview.List, workloadMenu *tview.Table) *tview.Table {

	menu.SetBorder(true).SetTitle("Settings")

	ix := nodes.GetCurrentItem()
	nodeSelection, _ := nodes.GetItemText(ix)

	cli.setWorkloadsSettings(nodeSelection, menu)

	menu.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEscape:
			app.SetFocus(workloads)
		}
	})

	return menu
}

func (cli *CLI) initWorkloadSettingsMenu(pages *tview.Pages, workloadMenu, workloads *tview.Table, menu, nodes, box *tview.List) *tview.Table {

	workloadMenu.SetBorder(true).SetTitle("Settings")

	ix := menu.GetCurrentItem()
	workloadSelection, _ := menu.GetItemText(ix)

	ix = nodes.GetCurrentItem()
	nodeSelection, _ := nodes.GetItemText(ix)

	workloads.Select(1, 1)

	r, _ := workloads.GetSelection()
	workloadNamespaceSelection := workloads.GetCell(r, 1)
	cli.setWorkloadSettings(workloadSelection, workloadNamespaceSelection.Text, nodeSelection, workloadMenu, workloads)

	workloadMenu.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEscape:
			pages.SwitchToPage(finderPage)
			app.SetFocus(workloads)
		}
	})

	workloadMenu.SetSelectedFunc(func(row int, column int) {
		pages.ShowPage("modal")
		app.SetFocus(box)
	})

	workloads.SetSelectedFunc(func(row int, column int) {

		ix := menu.GetCurrentItem()
		workloadSelection, _ := menu.GetItemText(ix)

		ix = nodes.GetCurrentItem()
		nodeSelection, _ := nodes.GetItemText(ix)

		r, _ := workloads.GetSelection()
		nsCell := workloads.GetCell(r, 1)

		cli.setWorkloadSettings(workloadSelection, nsCell.Text, nodeSelection, workloadMenu, workloads)

		pages.SwitchToPage(finderWorkloadPage)
		workloadMenu.Select(1, 0).SetFixed(1, 1).SetSelectable(true, false)
		app.SetFocus(workloadMenu)
	})

	return workloadMenu

}
