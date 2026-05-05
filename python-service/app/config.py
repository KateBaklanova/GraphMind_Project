import os

VECTOR_STORE_PATH = "data/faiss_index"
DOCUMENTS_PATH = "data/documents"
EMBEDDING_MODEL = "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
LLM_MODEL = "mistral"
CHROMA_DB_PATH = "chroma_db"
HOST = "0.0.0.0"
PORT = 8000
OLLAMA_URL = os.getenv("OLLAMA_URL", "http://localhost:11434") 
# OLLAMA_URL = os.getenv("OLLAMA_URL", "http://ollama:11434")