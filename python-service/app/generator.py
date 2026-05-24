import requests
import json
import re
from config import LLM_MODEL, OLLAMA_URL
from collections import Counter



def normalize_graph(data):

    VALID_TYPES = {
        "person",
        "place",
        "organization",
        "event",
        "work",
        "date",
        "concept"
    }

    BAD_LABELS = VALID_TYPES | {
        "entity",
        "object",
        "node",
        "thing",
        "item"
    }

    raw_nodes = data.get("nodes", [])
    raw_edges = data.get("edges", [])

    clean_nodes = []
    clean_edges = []

    node_ids = set()
    seen_nodes = set()
    seen_edges = set()


    # УЗЛЫ
    for n in raw_nodes:

        if not isinstance(n, dict):
            continue

        node_id = n.get("id", "")
        label = n.get("label", "")
        node_type = n.get("type", "concept")

        if isinstance(node_id, list):
            node_id = " ".join(map(str, node_id))

        if isinstance(label, list):
            label = " ".join(map(str, label))

        if isinstance(node_type, list):
            node_type = node_type[0] if node_type else "concept"

        raw_id = str(node_id).strip()

        if not raw_id:
            continue

        canonical_id = raw_id.lower()

        # дубли
        if canonical_id in seen_nodes:
            continue

        seen_nodes.add(canonical_id)


        node_type = str(node_type).strip().lower()

        if node_type not in VALID_TYPES:
            node_type = "concept"


        label = str(label).strip()

        if not label:
            label = raw_id

        if label.lower() in BAD_LABELS:
            label = raw_id

        clean_node = {
            "id": canonical_id,
            "label": label,
            "type": node_type
        }

        clean_nodes.append(clean_node)
        node_ids.add(canonical_id)

    # ДЛЯ РЕБЕР
    for e in raw_edges:

        if not isinstance(e, dict):
            continue

        from_id = e.get("from", "")
        to_id = e.get("to", "")
        label = e.get("label", "")

        # строка
        if isinstance(from_id, list):
            from_id = " ".join(map(str, from_id))

        if isinstance(to_id, list):
            to_id = " ".join(map(str, to_id))

        if isinstance(label, list):
            label = " ".join(map(str, label))

        from_id = str(from_id).strip().lower()
        to_id = str(to_id).strip().lower()
        
        label = str(label).strip().lower()

        if not from_id or not to_id:
            continue

        # self-loops
        if from_id == to_id:
            continue

        # существующие
        if from_id not in node_ids:
            continue

        if to_id not in node_ids:
            continue

        # плохие label
        if not label:
            continue

        if label in {
            "related",
            "connected",
            "associated",
            "linked",
            "relation"
        }:
            continue

        edge_key = (from_id, to_id, label)

        # deduplicate edges
        if edge_key in seen_edges:
            continue

        seen_edges.add(edge_key)

        clean_edges.append({
            "from": from_id,
            "to": to_id,
            "label": label
        })

    # ORPHAN 
    connected_nodes = set()

    for e in clean_edges:
        connected_nodes.add(e["from"])
        connected_nodes.add(e["to"])

    clean_nodes = [
        n for n in clean_nodes
        if n["id"] in connected_nodes
    ]

    return {
        "nodes": clean_nodes,
        "edges": clean_edges
    }


def extract_json(raw_output: str):

    raw_output = re.sub(r"```json|```", "", raw_output)

    match = re.search(r'\{[\s\S]*?\}', raw_output)
    if not match:
        print("NO JSON FOUND")
        print(repr(raw_output))
        return None

    cleaned = match.group()

    cleaned = cleaned.encode("utf-8", "ignore").decode("utf-8")

    return cleaned

def enrich_graph(graph):
    degree = Counter()

    for e in graph["edges"]:
        degree[e["from"]] += 1
        degree[e["to"]] += 1

    for n in graph["nodes"]:
        n["weight"] = degree[n["id"]]

    return graph


def build_extraction_chain():
    def extract(text):
        prompt = f"""
Extract a knowledge graph ONLY from the provided text.

Return ONLY valid JSON:
{{
  "nodes": [
    {{
      "id": "...",
      "label": "...",
      "type": "person|place|organization|event|work|date|concept"
    }}
  ],
  "edges": [
    {{
      "from": "...",
      "to": "...",
      "label": "..."
    }}
  ]
}}

STRICT RULES:
- Use ONLY entities explicitly mentioned in the text
- Do NOT infer missing entities
- Do NOT add background knowledge
- Do NOT complete partial facts
- If a relation is not explicit, do not create it
- Every edge must be directly supported by the text
- Do not invent nodes.
- Do not invent edges.
- Use only entities explicitly present in the text.
- Edges must reference existing node IDs only.
- If unsure return empty arrays.
- Prefer fewer edges over hallucinated edges
- IMPORTANT: label and type can NOT be the same. 

GRAPH RULES:
- Build cross-links when explicitly supported
- Avoid star graphs
- No relation nodes
- Use specific edge labels
- Never use "related"

NODE RULES:
- label = exact entity text from input
- type ∈ {{
  person, place, organization,
  event, work, date, concept
}}
- Never use type as label

OUTPUT:
- Return JSON only
- No explanations
- No markdown

Text:
"""+ text

        response = requests.post(
            f"{OLLAMA_URL}/api/generate",
            json={
                "model": LLM_MODEL,
                "prompt": prompt,
                "stream": False,
                "temperature": 0
            },
            timeout=600
        )

        if response.status_code != 200:
            raise Exception(f"Ollama error: {response.status_code}")

        result = response.json()
        raw_output = result.get("response", "")

        print("RAW:", raw_output)

        start = raw_output.find("{")
        end = raw_output.rfind("}") + 1

        if start == -1 or end == -1:
            print("NO JSON FOUND")
            return {"nodes": [], "edges": []}

        cleaned = raw_output[start:end]

        try:
            data = json.loads(cleaned)
        except Exception as e:
            print("JSON ERROR:", e)
            print("CLEANED:", repr(cleaned))
            return {"nodes": [], "edges": []}

        return enrich_graph(normalize_graph(data))

    return extract