package cli

import (
	"io/ioutil"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/ghodss/yaml"
	"github.com/rivo/tview"
	"github.com/soer3n/incident-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

const (
	batchSize          = 80                // The number of rows loaded per batch.
	finderPage         = "workloads"       // The name of the Finder page.
	finderHostPage     = "host"            // The name of the Finder page.
	finderWorkloadPage = "workloadsfinder" // The name of the Finder page.
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

	menu := tview.NewTable().SetBorders(false)
	menu.SetBorder(true).SetTitle("Settings")

	hostMenu := tview.NewTable().SetBorders(false)
	hostMenu.SetBorder(true).SetTitle("Settings")

	workloadMenu := tview.NewTable().SetBorders(false)
	workloadMenu.SetBorder(true).SetTitle("Settings")

	workloads := tview.NewList()
	workloads.SetBorder(true).SetTitle("Menu")

	nodes = cli.initMainMenu(nodes, workloads, menu, hostMenu)
	workloads = cli.initNodeMenu(nodes, workloads, menu, hostMenu)

	workloadBoxFunc := cli.initModalSelectionFunc()
	workloadBox := cli.initModalSelectionBox("modal", finderWorkloadPage, []string{"true", "false"}, workloadMenu, menu, nodes, workloads)

	hostBoxFunc := cli.initModalSelectionFunc()
	hostBox := cli.initModalSelectionBox("modalHost", finderHostPage, []string{"true", "false"}, hostMenu, menu, nodes, workloads)

	hostMenu = cli.initHostSettingsMenu(hostMenu, nodes, workloads, hostBox)

	workloadMenu = cli.initWorkloadSettingsMenu(workloadMenu, menu, workloads, nodes, workloadBox)
	menu = cli.initWorkloadsMenu(menu, nodes, workloads, workloadMenu)

	flexHost := tview.NewFlex().
		AddItem(nodes, 0, 1, true).
		AddItem(workloads, 0, 1, false).
		AddItem(hostMenu, 0, 3, false)
	pages = tview.NewPages().
		AddPage(finderHostPage, flexHost, true, true).
		AddPage("modalHost", hostBoxFunc(hostBox, 40, 20), true, false)

	flexWorkloads := tview.NewFlex().
		AddItem(nodes, 0, 1, true).
		AddItem(workloads, 0, 1, false).
		AddItem(workloadMenu, 0, 3, false)
	pages = pages.
		AddPage(finderWorkloadPage, flexWorkloads, true, false).
		AddPage("modal", workloadBoxFunc(workloadBox, 40, 20), true, false)

	flex := tview.NewFlex().
		AddItem(nodes, 0, 1, true).
		AddItem(workloads, 0, 1, false).
		AddItem(menu, 0, 3, false)

	pages = pages.
		AddPage(finderPage, flex, true, false)

	app.SetRoot(pages, true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			if err := cli.writeQuarantineSpec(); err != nil {
				cli.logger.Fatal(err)
			}
		}
		return event
	})
	return app, nil
}

func (cli *CLI) RenderRunView() (*tview.Application, error) {
	app = tview.NewApplication()
	return app, nil
}

func (cli CLI) writeQuarantineSpec() error {
	scheme := runtime.NewScheme()
	v1alpha1.AddToScheme(scheme)
	codec := serializer.NewCodecFactory(scheme).LegacyCodec(v1alpha1.GroupVersion)
	output, _ := runtime.Encode(codec, cli.q)

	res, err := yaml.JSONToYAML(output)

	if err != nil {
		return err
	}

	cd, err := os.Getwd()

	if err != nil {
		return err
	}

	if err = os.Chmod(cd, 0777); err != nil {
		return err
	}

	err = ioutil.WriteFile(cd+"/quarantine.yaml", res, 0644)

	if err != nil {
		return err
	}

	cli.logger.Info("data written")
	return nil
}
