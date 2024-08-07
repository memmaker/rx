package game

import (
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/recfile"
	"os"
	"strings"
)

type Conversation struct {
	openingBranches []OpeningBranch
	nodes           map[string]ConversationNode
}

func NewConversation() *Conversation {
	return &Conversation{nodes: make(map[string]ConversationNode)}
}

func (c *Conversation) GetRootNode() ConversationNode {
	for _, branch := range c.openingBranches {
		evaluateResult, err := branch.branchCondition.Evaluate(nil)
		asBool := evaluateResult.(bool)
		if err == nil && asBool {
			return c.nodes[branch.branchName]
		}
	}
	return ConversationNode{NpcText: "No opening branch found", Name: "invalid"}
}

func (c *Conversation) GetNextNode(chosenOption ConversationOption) ConversationNode {
	return c.nodes[chosenOption.GetFollowupBranch()]
}

func (c *Conversation) GetNodeByName(node string) ConversationNode {
	return c.nodes[node]
}

type OpeningBranch struct {
	branchCondition *govaluate.EvaluableExpression
	branchName      string
}

type ConversationNode struct {
	Name    string
	NpcText string
	Effects []string
	Options []ConversationOption
}

type ConversationOption struct {
	displayCondition *govaluate.EvaluableExpression
	playerText       string
	branchCondition  *govaluate.EvaluableExpression
	successBranch    string // will default to the current node if not set
	failureBranch    string
}

func (o *ConversationOption) CanDisplay() bool {
	if o.displayCondition == nil {
		return true
	}
	evaluateResult, err := o.displayCondition.Evaluate(nil)
	asBool := evaluateResult.(bool)
	return err == nil && asBool
}

func (o *ConversationOption) GetFollowupBranch() string {
	if o.branchCondition == nil {
		return o.successBranch
	}
	evaluateResult, err := o.branchCondition.Evaluate(nil)
	asBool := evaluateResult.(bool)
	if err == nil && asBool {
		return o.successBranch
	}
	return o.failureBranch
}

func (g *GameState) ParseConversation(filename string) (*Conversation, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	flags := g.gameFlags

	conditionFuncs := map[string]govaluate.ExpressionFunction{
		"HasFlag": func(args ...interface{}) (interface{}, error) {
			flagName := args[0].(string)
			return flags.HasFlag(flagName), nil
		},
	}

	records := recfile.ReadMulti(file)
	conversation := NewConversation()
	openingBranches := make([]OpeningBranch, 0)
	for _, branchRecords := range records["OpeningBranch"] {
		var branch OpeningBranch
		for _, fields := range branchRecords {
			if fields.Name == "cond" {
				branch.branchCondition, _ = govaluate.NewEvaluableExpressionWithFunctions(fields.Value, conditionFuncs)
			} else if fields.Name == "goto" {
				branch.branchName = fields.Value
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
			if field.Name == "name" {
				conversationNode.Name = field.Value
			} else if field.Name == "npc" {
				conversationNode.NpcText = field.Value
			} else if field.Name == "effect" {
				conversationNode.Effects = append(conversationNode.Effects, field.Value)
			} else if strings.HasPrefix(field.Name, "o_") {
				if field.Name == "o_text" {
					if currentOption.playerText != "" {
						conversationNode.Options = append(conversationNode.Options, currentOption)
					}
					currentOption.playerText = field.Value
					currentOption.branchCondition = nil
					currentOption.successBranch = ""
					currentOption.failureBranch = ""
					currentOption.displayCondition = nil
				} else if field.Name == "o_cond" {
					currentOption.displayCondition, _ = govaluate.NewEvaluableExpressionWithFunctions(field.Value, conditionFuncs)
				} else if field.Name == "o_goto" || field.Name == "o_succ" {
					currentOption.successBranch = field.Value
				} else if field.Name == "o_fail" {
					currentOption.failureBranch = field.Value
				} else if field.Name == "o_test" {
					currentOption.branchCondition, _ = govaluate.NewEvaluableExpressionWithFunctions(field.Value, conditionFuncs)
				}
			}
		}
		if currentOption.playerText != "" {
			conversationNode.Options = append(conversationNode.Options, currentOption)
			currentOption.playerText = ""
			currentOption.branchCondition = nil
			currentOption.successBranch = ""
			currentOption.failureBranch = ""
			currentOption.displayCondition = nil
		}
		allNodes[conversationNode.Name] = conversationNode
	}
	conversation.nodes = allNodes
	return conversation, nil
}
