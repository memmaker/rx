#!/bin/zsh
DIALOGUE="$1"
./contractor graph_dialogue DefaultIntroduction "${DIALOGUE}.rec" | dot -Tpng > digraph.png && open digraph.png