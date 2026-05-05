import networkx as nx

def build_graph(relations):
    G = nx.DiGraph()

    for r in relations:
        subject = r.get("subject")
        relation = r.get("relation")
        obj = r.get("object")
        
        if subject and relation and obj:
            G.add_node(subject)
            G.add_node(obj)
            G.add_edge(subject, obj, label=relation)

    return G