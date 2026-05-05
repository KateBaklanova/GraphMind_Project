FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    gcc g++ \
    && rm -rf /var/lib/apt/lists/*

COPY python-service/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY python-service/app/ ./app/
COPY python-service/pipeline/ ./pipeline/
COPY python-service/visualization/ ./visualization/

RUN mkdir -p static data

EXPOSE 8000

CMD ["python", "-m", "app.main"]