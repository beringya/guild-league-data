from fastapi import APIRouter

from app.api.routes import auth, avatars, backups, battles, rankings, settings

api_router = APIRouter(prefix="/api")
api_router.include_router(auth.router)
api_router.include_router(battles.router)
api_router.include_router(rankings.router)
api_router.include_router(settings.router)
api_router.include_router(backups.router)
api_router.include_router(avatars.router, prefix="")
