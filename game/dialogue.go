package game

import (
	"contractor/d100"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/recfile"
	"os"
	"regexp"
	"strings"
)

type Conversation struct {
	openingBranches []OpeningBranch
	nodes           map[string]ConversationNode
	Variables       map[string]interface{}
}

func NewConversation() *Conversation {
	return &Conversation{nodes: make(map[string]ConversationNode)}
}

func (c *Conversation) CreateGraph() string {
	// output the conversation graph in the graphviz dot format
	graph := "digraph G {\n"
	for nodeName, node := range c.nodes {
		graph += nodeName + " [label=\"" + node.NpcText + "\"];\n"
		for _, option := range node.Options {
			for _, branch := range option.GetAllPossibleBranches() {
				graph += nodeName + " -> " + branch + " [label=\"" + option.playerText + "\"];\n"
			}

		}

	}
	graph += "}"
	return graph
}
func (c *Conversation) SetVariables(params map[string]interface{}) {
	c.Variables = params
}
func (c *Conversation) MergeVariables(params map[string]interface{}) {
	if c.Variables == nil {
		c.Variables = params
		return
	}
	for key, value := range params {
		c.Variables[key] = value
	}
}
func (c *Conversation) GetRootNode() ConversationNode {
	for _, branch := range c.openingBranches {
		if branch.Name != "" { // named branches are used for NPC initiated conversations
			continue
		}
		evaluateResult, err := branch.BranchCondition.Evaluate(c.Variables)
		asBool := evaluateResult.(bool)
		if err == nil && asBool {
			return c.nodes[branch.BranchName]
		}
	}
	return ConversationNode{NpcText: "No opening branch found", Name: "invalid"}
}

func (c *Conversation) GetOpeningBranches() []OpeningBranch {
	return c.openingBranches
}

func (c *Conversation) GetNextNode(chosenOption ConversationOption) ConversationNode {
	return c.nodes[chosenOption.GetFollowupBranch(c.Variables)]
}

func (c *Conversation) GetNodeByName(node string) ConversationNode {
	return c.nodes[node]
}

func (c *Conversation) GetAllNodes() map[string]ConversationNode {
	return c.nodes
}

func (c *Conversation) GetOpeningBranchByName(name string) OpeningBranch {
	for _, branch := range c.openingBranches {
		if branch.Name == name {
			return branch
		}
	}
	return OpeningBranch{}
}

type OpeningBranch struct {
	Name            string
	BranchCondition *govaluate.EvaluableExpression
	BranchName      string
}

type ConversationNode struct {
	Name    string
	NpcText string
	Effects []string
	Options []ConversationOption
}

func (n *ConversationNode) IsEmpty() bool {
	return n.Name == "" && n.NpcText == "" && len(n.Options) == 0
}

type ConversationOption struct {
	displayCondition *govaluate.EvaluableExpression
	playerText       string
	branchCondition  *govaluate.EvaluableExpression
	successBranch    string // will default to the current node if not set
	failureBranch    string
	effects          []string
}

func (o *ConversationOption) CanDisplay(params map[string]interface{}) bool {
	if o.displayCondition == nil {
		return true
	}
	evaluateResult, err := o.displayCondition.Evaluate(params)
	if err != nil {
		panic(err)
	}
	asBool := evaluateResult.(bool)
	return err == nil && asBool
}

func (o *ConversationOption) GetFollowupBranch(params map[string]interface{}) string {
	if o.branchCondition == nil {
		return o.successBranch
	}
	evaluateResult, err := o.branchCondition.Evaluate(params)
	asBool := evaluateResult.(bool)
	if err == nil && asBool {
		return o.successBranch
	}
	return o.failureBranch
}

func (o *ConversationOption) GetAllPossibleBranches() []string {
	if o.branchCondition == nil {
		return []string{o.successBranch}
	}
	return []string{o.successBranch, o.failureBranch}
}

func (o *ConversationOption) RollInfo() string {
	if o.branchCondition == nil {
		return ""
	}
	conditionAsString := o.branchCondition.String()
	// format : RollSkill('skillName', -10)
	regexPattern := `RollSkill\('([^']+)',\s*'([^']+)'\)`
	// extract skill name
	// extract modifier
	matches := regexp.MustCompile(regexPattern).FindStringSubmatch(conditionAsString)
	if matches == nil {
		return ""
	}
	skillName := d100.SkillFromString(matches[1])
	difficulty := matches[2]
	return fmt.Sprintf(" (%s - %s)", skillName.String(), difficulty)
}

func (o *ConversationOption) GetDisplayCondition() *govaluate.EvaluableExpression {
	return o.displayCondition
}

func (o *ConversationOption) GetEffects() []string {
	return o.effects
}

func (o *ConversationOption) GetBranchCondition() *govaluate.EvaluableExpression {
	return o.branchCondition
}

func (o *ConversationOption) GetSuccessBranch() string {
	return o.successBranch
}

func (o *ConversationOption) GetFailureBranch() string {
	return o.failureBranch
}

func (o *ConversationOption) GetGotoBranch() string {
	return o.successBranch
}

func ParseConversation(filename string, conditionFuncs map[string]govaluate.ExpressionFunction) (*Conversation, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	records, _ := recfile.ReadMulti(file)
	conversation := NewConversation()
	openingBranches := make([]OpeningBranch, 0)
	variables := make(map[string]interface{})

	if records["Variables"] != nil {
		for _, variableRecord := range records["Variables"] {
			for _, field := range variableRecord {
				expression, _ := govaluate.NewEvaluableExpressionWithFunctions(field.Value, conditionFuncs)
				variables[field.Name], _ = expression.Evaluate(nil)
			}
		}
		conversation.SetVariables(variables)
	}

	for _, branchRecords := range records["OpeningBranch"] {
		var branch OpeningBranch
		for _, fields := range branchRecords {
			fieldName := strings.ToLower(fields.Name)
			if fieldName == "cond" {
				cond, parseErr := govaluate.NewEvaluableExpressionWithFunctions(fields.Value, conditionFuncs)
				if parseErr != nil {
					panic(parseErr)
				}
				branch.BranchCondition = cond
			} else if fieldName == "goto" {
				branch.BranchName = fields.Value
			} else if fieldName == "name" {
				branch.Name = fields.Value
			}
		}
		openingBranches = append(openingBranches, branch)
	}
	conversation.openingBranches = openingBranches
	allNodes := make(map[string]ConversationNode)
	for _, nodeRecord := range records["Nodes"] {
		var conversationNode ConversationNode
		var currentOption ConversationOption
		for _, field := range nodeRecord {
			fieldName := strings.ToLower(field.Name)
			if fieldName == "name" {
				conversationNode.Name = field.Value
			} else if fieldName == "npc" {
				conversationNode.NpcText = strings.TrimSpace(field.Value)
			} else if fieldName == "effect" {
				conversationNode.Effects = append(conversationNode.Effects, field.Value)
			} else if strings.HasPrefix(fieldName, "o_") {
				if fieldName == "o_text" {
					if currentOption.playerText != "" {
						conversationNode.Options = append(conversationNode.Options, currentOption)
					}
					currentOption.playerText = strings.TrimSpace(field.Value)
					currentOption.branchCondition = nil
					currentOption.successBranch = ""
					currentOption.failureBranch = ""
					currentOption.displayCondition = nil
					currentOption.effects = nil
				} else if fieldName == "o_cond" {
					currentOption.displayCondition, _ = govaluate.NewEvaluableExpressionWithFunctions(field.Value, conditionFuncs)
				} else if fieldName == "o_goto" || fieldName == "o_succ" {
					currentOption.successBranch = field.Value
				} else if fieldName == "o_effect" {
					currentOption.effects = append(currentOption.effects, field.Value)
				} else if fieldName == "o_fail" {
					currentOption.failureBranch = field.Value
				} else if fieldName == "o_test" {
					currentOption.branchCondition, _ = govaluate.NewEvaluableExpressionWithFunctions(field.Value, conditionFuncs)
				}
			}
		}
		if currentOption.playerText != "" {
			conversationNode.Options = append(conversationNode.Options, currentOption)
		}
		allNodes[conversationNode.Name] = conversationNode
	}
	conversation.nodes = allNodes
	return conversation, nil
}
