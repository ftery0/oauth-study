# app3 HelpDesk backend — FastAPI + PostgreSQL
FROM python:3.12-slim
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential libpq-dev && rm -rf /var/lib/apt/lists/*

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY main.py ./
COPY app ./app

EXPOSE 8013
ENV PORT=8013 HOST=0.0.0.0
CMD ["python","main.py"]
