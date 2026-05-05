from pyvis.network import Network

def visualize_graph(G, output="graph.html"):
    net = Network(
        height="800px",
        width="100%",
        bgcolor="#111111",
        font_color="white",
        directed=True
    )

    # Ноды
    for node_id, data in G.nodes(data=True):
        weight = data.get("weight", 1)

        if weight >= 6:
            importance = "high"
        elif weight >= 3:
            importance = "medium"
        else:
            importance = "low"
            
        print(importance)

        color = {
            "high": "#ff4d4d",     
            "medium": "#4da6ff",   
            "low": "#888888"       
        }.get(importance, "#aaaaaa")

        size = {
            "high": 35,
            "medium": 22,
            "low": 12
        }.get(importance, 18)

        title = f"""
        <b>{data.get("label", node_id)}</b><br>
        type: {data.get("type", "")}<br>
        importance: {importance}<br>
        {data.get("description", "")}
        """

        net.add_node(
            node_id,
            label=data.get("label", node_id),  
            title=title,
            color=color,
            size=size
        )

    # Ребра
    for u, v, data in G.edges(data=True):
        label = data.get("label") or "related"

        net.add_edge(
            u,
            v,
            label=label,
            title=label,
            color="#cccccc"
        )

    # Фищика
    net.barnes_hut()
    net.repulsion(node_distance=200, central_gravity=0.3)

    net.show_buttons(filter_=['physics'])
    net.write_html(output)

    print(f"Graph saved to {output}")