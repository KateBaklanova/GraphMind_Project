import sys
import os
from pathlib import Path

sys.path.append(str(Path(__file__).parent.parent))

from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
import uvicorn
import logging
from generator import build_extraction_chain
from config import HOST, PORT

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="Knowledge Graph Extractor")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

extract = build_extraction_chain()

@app.post("/visualize")
async def visualize(request: Request):
    try:
        data = await request.json()
        text = data.get("text")
        
        if not text:
            return {"error": "text required"}
        
        graph_data = extract(text)
        
        return {
            "message": "Граф создан",
            "triplets_count": len(graph_data.get("edges", [])),
            "nodes_count": len(graph_data.get("nodes", [])),
            "edges_count": len(graph_data.get("edges", [])),
            "graph_data": graph_data
        }
        
    except Exception as e:
        logger.error(f"Ошибка: {e}")
        return {"error": str(e)}

@app.get("/")
async def root():
    return {
        "message": "Knowledge Graph Extractor",
        "endpoint": "/visualize - POST с текстом"
    }

if __name__ == "__main__":
    uvicorn.run(app, host=HOST, port=PORT)