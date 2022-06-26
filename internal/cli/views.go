package cli

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ghodss/yaml"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/soer3n/incident-operator/api/v1alpha1"
	"github.com/soer3n/incident-operator/internal/templates/loader"
	"github.com/soer3n/incident-operator/internal/utils"
	"github.com/soer3n/incident-operator/webhooks/quarantine"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	finderPage         = "workloads"       // The name of the Finder page.
	finderHostPage     = "host"            // The name of the Finder page.
	finderWorkloadPage = "workloadsfinder" // The name of the Finder page.
)

var (
	app *tview.Application // The tview application.
)

var (
	scheme = runtime.NewScheme()
)

func (cli *CLI) RenderPrepareView() (*tview.Application, error) {

	app = tview.NewApplication()
	pages := cli.getEditPages()

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
	runPage := cli.getRunPage()

	pages := tview.NewPages().AddPage("runMenu", runPage, true, true)

	app.SetRoot(pages, true)

	return app, nil
}

func (cli *CLI) getEditPages() *tview.Pages {

	pages := tview.NewPages()

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
	workloads = cli.initNodeMenu(pages, nodes, workloads, menu, hostMenu)

	menu = cli.initWorkloadsMenu(pages, menu, nodes, workloads, workloadMenu)

	workloadBoxFunc := cli.initModalSelectionFunc()
	workloadBox := cli.initModalSelectionBox(pages, "modal", finderWorkloadPage, []string{"true", "false"}, workloadMenu, menu, nodes, workloads)

	workloadMenu = cli.initWorkloadSettingsMenu(pages, workloadMenu, menu, workloads, nodes, workloadBox)

	hostBoxFunc := cli.initModalSelectionFunc()
	hostBox := cli.initModalSelectionBox(pages, "modalHost", finderHostPage, []string{"true", "false"}, hostMenu, menu, nodes, workloads)

	hostMenu = cli.initHostSettingsMenu(pages, hostMenu, nodes, workloads, hostBox)

	flexHost := tview.NewFlex().
		AddItem(nodes, 0, 1, true).
		AddItem(workloads, 0, 1, false).
		AddItem(hostMenu, 0, 3, false)

	pages.
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

	return pages
}

func (cli *CLI) getRunPage() tview.Primitive {

	menu := tview.NewTreeNode("overview").
		SetColor(tcell.ColorRed)
	tree := tview.NewTreeView().
		SetRoot(menu).
		SetCurrentNode(menu)

	nodeObjs := cli.getNodes()

	add := func(target *tview.TreeNode, node string) {
		pods := cli.getPodsByNode("", node)
		for _, pod := range pods.Items {
			node := tview.NewTreeNode(pod.GetName()).
				SetReference(pod.GetName()).
				SetSelectable(true)
			node.SetColor(tcell.ColorGreen)

			target.AddChild(node)
		}
	}

	for _, node := range nodeObjs.Items {
		nodeObj := tview.NewTreeNode(node.Name).
			SetReference(node.Name).
			SetSelectable(true)
		nodeObj.SetColor(tcell.ColorGreen)

		add(nodeObj, node.Name)
		nodeObj.Collapse()
		nodeObj.SetSelectedFunc(func() {
			nodeObj.Expand()
		})
		menu.AddChild(nodeObj)
	}

	events := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).SetScrollable(true)

	c := utils.GetTypedKubernetesClient()

	listOpts := metav1.ListOptions{
		LabelSelector: "control-plane=controller-manager",
	}

	pods, err := c.CoreV1().Pods("incident-operator-system").List(context.Background(), listOpts)

	if err != nil {
		cli.logger.Fatal(err.Error())
	}

	ctx := context.TODO()
	// ctx, _ = context.WithCancel(ctx)
	channel := make(chan string, 1)

	for _, pod := range pods.Items {
		go getControllerLogs(channel, c, ctx, pod.Name, pod.Namespace)
	}

	fmt.Fprint(events, "[green]Init: [white]starting event and log streaming...\n")

	go handleQuaratineController(ctx, channel)

	go func(ctx context.Context) {
		for line := range channel {
			fmt.Fprint(events, line)
		}
	}(ctx)

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRight:
			app.SetFocus(events)
		}
		return event
	})

	events.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			app.SetFocus(tree)
		}
		return event
	})

	time.Sleep(time.Second * 2)

	flex := tview.NewFlex().
		AddItem(tree, 0, 1, true).
		AddItem(events, 0, 2, false)

	return flex
}

func handleQuaratineController(ctx context.Context, channel chan string) {
	var dec *admission.Decoder
	var err error

	if dec, err = admission.NewDecoder(scheme); err != nil {
		channel <- err.Error()
		return
	}

	cl, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		channel <- "failed to create client"
		return
	}

	h := &quarantine.QuarantineValidateHandler{
		Decoder: dec,
		Client:  cl,
		Log:     ctrl.Log.WithName("validation").WithName("quarantine-controller"),
	}

	currentDir, _ := os.Getwd()
	spec, err := loader.LoadQuarantineSpec(currentDir+"/quarantine.yaml", logrus.StandardLogger())

	if err != nil {
		channel <- err.Error()
		return
	}

	if err = h.Validate(spec); err != nil {
		channel <- err.Error()
		channel <- "rescheduling controller..."
		if err = RescheduleQuarantineController([]string{}); err != nil {
			channel <- err.Error()
			return
		}
		return
	}
}

func getControllerLogs(channel chan string, c *kubernetes.Clientset, ctx context.Context, name, namespace string) {

	logStream, err := c.CoreV1().Pods(namespace).GetLogs(name, &v1.PodLogOptions{Follow: true, Container: "manager"}).Stream(ctx)

	if err != nil {
		channel <- "[red]exiting due to error:[white]" + err.Error()
		return
	}

	defer logStream.Close()
	reader := bufio.NewScanner(logStream)
	var line string

	for {
		for reader.Scan() {
			select {
			case <-ctx.Done():
				break
			default:
				line = reader.Text()
				channel <- fmt.Sprintf("[yellow]Controller Logs:[white] %v\n", line)
			}
		}
	}
}

func (cli CLI) writeQuarantineSpec() error {
	scheme := runtime.NewScheme()

	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return err
	}

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
