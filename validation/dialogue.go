package validation

import (
	"contractor/game"
	"contractor/util"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/fxtools"
	"html"
	"os"
	"path"
	"slices"
	"strings"
)

// We want a dialogue script checker
// It should go through all dialogue files
// 1. Looking for missing nodes

type DialogueReport struct {
	DialogueFile string
	Nodes        map[string]NodeInfos
}

func (dr DialogueReport) FlagsQueried() []string {
	flags := make(map[string]bool)
	for _, node := range dr.Nodes {
		for flag := range node.FlagsQueried {
			flags[flag] = true
		}
	}
	asList := util.MapKeys(flags)
	slices.SortStableFunc(asList, strings.Compare)
	return asList
}

func (dr DialogueReport) FlagsSet() []string {
	flags := make(map[string]bool)
	for _, node := range dr.Nodes {
		for flag := range node.FlagsSet {
			flags[flag] = true
		}
	}
	asList := util.MapKeys(flags)
	slices.SortStableFunc(asList, strings.Compare)
	return asList
}

func (dr DialogueReport) MissingNodes() []string {
	var missing []string
	for name := range dr.NodeNamesMentioned() {
		if _, ok := dr.Nodes[name]; !ok {
			missing = append(missing, name)
		}
	}

	slices.SortStableFunc(missing, strings.Compare)

	return missing
}

func (dr DialogueReport) NodeNames() []string {
	names := make([]string, 0, len(dr.Nodes))
	for name := range dr.Nodes {
		names = append(names, name)
	}
	slices.SortStableFunc(names, strings.Compare)
	return names
}

func (dr DialogueReport) UnreferencedNodes() []string {
	var unreferenced []string
	mentioned := dr.NodeNamesMentioned()
	for name := range dr.Nodes {
		if _, ok := mentioned[name]; !ok {
			unreferenced = append(unreferenced, name)
		}
	}

	slices.SortStableFunc(unreferenced, strings.Compare)

	return unreferenced
}

func (dr DialogueReport) AsDotGraph(annotateEdges bool, hideBackLinksToNode string) string {
	endNodes := make(map[string]bool)
	startNodes := make(map[string]bool)
	isStartNode := func(name string) bool {
		_, ok := startNodes[name]
		return ok
	}
	rootNode := dr.Nodes["START"]
	for _, mentioned := range rootNode.NamesMentionedAsList() {
		startNodes[mentioned] = true
	}
	var builder strings.Builder
	builder.WriteString("digraph Dialogue {\n")
	for _, nodeName := range dr.NodeNames() {
		node := dr.Nodes[nodeName]

		if len(node.FlagsSet) > 0 {
			/*
				[label=<Birth of George Washington<BR />
				        <FONT POINT-SIZE="10">See also: American Revolution</FONT>>];
			*/
			builder.WriteString("  \"")
			builder.WriteString(nodeName)
			builder.WriteString("\" [label=<")
			builder.WriteString(nodeName)
			builder.WriteString("<BR /><FONT POINT-SIZE=\"10\">")
			for flag := range node.FlagsSet {
				builder.WriteString(flag)
				builder.WriteString("<BR />")
			}
			builder.WriteString("</FONT>>];\n")
		}

		for _, mentioned := range node.NamesMentionedAsList() {
			cond := node.Transition[mentioned]
			if nodeName == "START" {
				// mentioned [color="green"]
				builder.WriteString("  \"")
				builder.WriteString(mentioned)
				builder.WriteString("\"")
				builder.WriteString(" [color=\"green\"")
				if cond != "" {
					builder.WriteString(", label=<")
					builder.WriteString(mentioned)
					builder.WriteString("<BR /><FONT POINT-SIZE=\"10\">")
					builder.WriteString(cond)
					builder.WriteString("<BR />")
					builder.WriteString("</FONT>>];\n")
				} else {
					builder.WriteString("];\n")
				}
				continue
			}

			if mentioned == hideBackLinksToNode && !isStartNode(nodeName) {
				continue
			}

			if strings.HasPrefix(strings.ToLower(mentioned), "end") {
				endNodes[mentioned] = true
			}

			builder.WriteString("  \"")
			builder.WriteString(nodeName)
			builder.WriteString("\" -> \"")
			builder.WriteString(mentioned)
			if cond != "" && annotateEdges {
				builder.WriteString("\" [label=\"")
				builder.WriteString(cond)
				builder.WriteString("\"];\n")
			} else {
				builder.WriteString("\";\n")
			}

		}
	}
	endNodeList := util.MapKeys(endNodes)
	slices.SortStableFunc(endNodeList, strings.Compare)
	for _, endNode := range endNodeList {
		builder.WriteString("  \"")
		builder.WriteString(endNode)
		builder.WriteString("\" [color=\"red\"];\n")
	}
	builder.WriteString("}\n")
	return builder.String()
}

func (dr DialogueReport) NodeNamesMentioned() map[string]bool {
	names := make(map[string]bool)
	for _, info := range dr.Nodes {
		for mentioned := range info.Transition {
			names[mentioned] = true
		}
	}
	return names
}

type DialogueChecker struct {
	rootDir string
	reports []DialogueReport
}

func GraphDialogue(rootDir string, scriptFuncs map[string]govaluate.ExpressionFunction, filename string, hideBackLinksToNode string) {
	checker := NewDialogueChecker(path.Join(rootDir, "dialogues"))
	report := checker.CheckSingleFile(filename, scriptFuncs)
	fmt.Println(report.AsDotGraph(false, hideBackLinksToNode))
}
func ValidateDialogue(rootDir string, scriptFuncs map[string]govaluate.ExpressionFunction) {
	checker := NewDialogueChecker(path.Join(rootDir, "dialogues"))
	reports := checker.CreateReports(scriptFuncs)
	for _, report := range reports {
		missing := report.MissingNodes()
		unreferenced := report.UnreferencedNodes()
		if len(missing) > 0 || len(unreferenced) > 0 || len(report.FlagsQueried()) > 0 || len(report.FlagsSet()) > 0 {
			println("\nDialogue file: ", report.DialogueFile)
			if len(missing) > 0 {
				println("Missing nodes: ")
				for _, name := range missing {
					println("  ", name)
				}
			}
			if len(unreferenced) > 0 {
				println("Unreferenced nodes: ")
				for _, name := range unreferenced {
					println("  ", name)
				}
			}
			if len(report.FlagsSet()) > 0 {
				println("Flags set: ")
				for flag := range report.FlagsSet() {
					println("  ", flag)
				}
			}
			if len(report.FlagsQueried()) > 0 {
				println("Flags queried: ")
				for _, flag := range report.FlagsQueried() {
					println("  ", flag)
				}
			}
		}
	}
}

func NewDialogueChecker(rootDir string) *DialogueChecker {
	return &DialogueChecker{rootDir: rootDir}
}

func (dc *DialogueChecker) CheckSingleFile(filename string, scriptFuncs map[string]govaluate.ExpressionFunction) DialogueReport {
	return DialogueReportFromFile(dc.rootDir, filename, scriptFuncs)
}

func (dc *DialogueChecker) CreateReports(scriptFuncs map[string]govaluate.ExpressionFunction) []DialogueReport {
	dir, err := os.ReadDir(dc.rootDir)
	if err != nil {
		return nil
	}
	var reports []DialogueReport
	for _, entry := range dir {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), "_") {
			continue
		}

		report := DialogueReportFromFile(dc.rootDir, entry.Name(), scriptFuncs)
		reports = append(reports, report)
	}
	return reports
}

func DialogueReportFromFile(directory string, filename string, scriptFuncs map[string]govaluate.ExpressionFunction) DialogueReport {
	filePath := path.Join(directory, filename)
	conv, _ := game.ParseConversation(filePath, scriptFuncs)

	rootNode := NodeInfos{
		Name:         "START",
		Transition:   make(map[string]string),
		FlagsSet:     make(map[string]bool),
		FlagsQueried: make(map[string]bool),
	}
	rootNode.Transition = appendFromOpeningBranch(rootNode.Transition, conv.GetOpeningBranches())
	allNodes := map[string]NodeInfos{
		rootNode.Name: rootNode,
	}
	allNodes = appendFromNodes(allNodes, conv.GetAllNodes())

	report := DialogueReport{
		DialogueFile: filename,
		Nodes:        allNodes,
	}
	return report
}

func appendFromOpeningBranch(mentioned map[string]string, nodes []game.OpeningBranch) map[string]string {
	for _, node := range nodes {
		cond := ""
		branchCondition := node.BranchCondition
		if branchCondition != nil {
			cond = branchCondition.String()
		}
		// & -> &amp;

		mentioned[node.BranchName] = html.EscapeString(cond)
		//cond := record.BranchCondition.String()
	}
	return mentioned
}

type NodeInfos struct {
	Name         string
	Transition   map[string]string
	FlagsSet     map[string]bool
	FlagsQueried map[string]bool
}

func (i NodeInfos) WithNamePresent(value string) NodeInfos {
	i.Name = value
	return i
}

func (i NodeInfos) WithTransitionTo(value string, cond string) NodeInfos {
	i.Transition[value] = html.EscapeString(cond)
	return i
}

func (i NodeInfos) WithFlagSet(flagname string) NodeInfos {
	i.FlagsSet[flagname] = true
	return i
}

func (i NodeInfos) WithFlagQueried(flagname string) NodeInfos {
	i.FlagsQueried[flagname] = true
	return i
}

func (i NodeInfos) NamesMentionedAsList() []string {
	keys := util.MapKeys(i.Transition)
	slices.SortStableFunc(keys, strings.Compare)
	return keys
}
func appendFromNodes(nodes map[string]NodeInfos, records map[string]game.ConversationNode) map[string]NodeInfos {
	for nodeName, node := range records {
		nextNode := NodeInfos{
			Transition:   make(map[string]string),
			FlagsSet:     make(map[string]bool),
			FlagsQueried: make(map[string]bool),
		}
		nextNode = nextNode.WithNamePresent(nodeName)
		for _, option := range node.Options {
			dispCondition := option.GetDisplayCondition()
			branchCondition := option.GetBranchCondition()
			dispCondString := ""
			if branchCondition == nil {
				if dispCondition != nil {
					dispCondString = dispCondition.String()
					nextNode = nextNode.WithTransitionTo(option.GetGotoBranch(), dispCondString)
				} else {
					nextNode = nextNode.WithTransitionTo(option.GetGotoBranch(), "")
				}
			} else {
				branchCondString := branchCondition.String()
				if dispCondition != nil {
					dispCondString = " " + dispCondition.String()
				}
				nextNode = nextNode.WithTransitionTo(option.GetSuccessBranch(), branchCondString+""+dispCondString)
				nextNode = nextNode.WithTransitionTo(option.GetFailureBranch(), branchCondString+""+dispCondString)
			}
			condition := option.GetDisplayCondition()
			if condition != nil {
				displayCondition := condition.String()
				hasflag := strings.Index(strings.ToLower(displayCondition), "hasflag")
				if hasflag != -1 {
					part := displayCondition[hasflag:]
					endIndex := strings.Index(part, "')") + 2
					flagname := part[:endIndex]
					if fxtools.LooksLikeAFunction(flagname) {
						_, args := fxtools.GetNameAndArgs(flagname)
						nextNode = nextNode.WithFlagQueried(args.Get(0))
					}
				}
			}
		}
		for _, effect := range node.Effects {
			if fxtools.LooksLikeAFunction(effect) {
				name, args := fxtools.GetNameAndArgs(effect)
				lowerFuncName := strings.ToLower(name)
				if lowerFuncName == "gotonode" {
					nextNode = nextNode.WithTransitionTo(args.Get(0), "")
				} else if lowerFuncName == "setflag" {
					nextNode = nextNode.WithFlagSet(args.Get(0))
				}
			}
		}
		nodes[nextNode.Name] = nextNode
	}
	return nodes
}
