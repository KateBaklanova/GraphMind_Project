import requests
import json
import re
from config import LLM_MODEL, OLLAMA_URL
from collections import Counter

# Подумать над добавление дифференциация по формам и цвету !!!
# Для создания архитектурных схем ???

def normalize_graph(data):
    nodes = data.get("nodes", [])
    edges = data.get("edges", [])

    node_ids = set()

    for n in nodes:
        n["id"] = n.get("id", "").strip().lower()
        node_ids.add(n["id"])

    for e in edges:
        e["from"] = e.get("from", "").strip().lower()
        e["to"] = e.get("to", "").strip().lower()

        if not e.get("label"):
            e["label"] = "related"

    return {"nodes": nodes, "edges": edges}


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
        prompt = """You are building a NON-TRIVIAL knowledge graph.

Return ONLY valid JSON.

FORMAT:
{"nodes": [{"id": "...", "label": "...", "type": "..."}],
 "edges": [{"from": "...", "to": "...", "label": "..."}]}

CRITICAL RULES:

1. DO NOT create a star graph (one central node connected to everything)

2. Build MULTI-LEVEL structure:
   - entities must connect to each other
   - not only to the main entity

3. Use CHAIN relationships:
   GOOD:
   author → wrote → book → set_in → place
   families → interact_with → each other

4. Create CROSS connections between entities

5. IMPORTANT:
   - central nodes should NOT dominate
   - distribute connections across graph

6. Add semantic diversity:
   - strong relations (wrote, belongs_to)
   - contextual (influenced_by, part_of, located_in)

7. Graph should look like a NETWORK, not a tree

8. Avoid generic labels like "related"

Return ONLY JSON. No text.

Text:
""" + text

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