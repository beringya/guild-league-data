from __future__ import annotations

import logging
from pathlib import Path

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import FileResponse
from fastapi.staticfiles import StaticFiles

from app.api.router import api_router
from app.core.bootstrap import bootstrap_database
from app.core.config import get_settings
from app.core.database import SessionLocal, engine

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s %(message)s")
settings = get_settings()

app = FastAPI(title=settings.app_name, version=settings.app_version)

if settings.cors_origins:
    app.add_middleware(
        CORSMiddleware,
        allow_origins=[origin.strip() for origin in settings.cors_origins.split(",") if origin.strip()],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )


@app.on_event("startup")
def startup() -> None:
    with SessionLocal() as db:
        bootstrap_database(db)


@app.get("/api/health")
def health() -> dict:
    with engine.connect() as conn:
        conn.exec_driver_sql("SELECT 1")
    return {"ok": True, "app_version": settings.app_version, "database": "ok"}


app.include_router(api_router)

static_dir = Path(__file__).resolve().parents[1] / "static"
if static_dir.exists():
    app.mount("/assets", StaticFiles(directory=static_dir / "assets"), name="assets")

    @app.get("/{path:path}")
    def spa_fallback(path: str) -> FileResponse:
        target = static_dir / path
        if target.is_file():
            return FileResponse(target)
        return FileResponse(static_dir / "index.html")
