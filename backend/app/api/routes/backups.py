from __future__ import annotations

from fastapi import APIRouter, Depends, HTTPException, status

from app.api.deps import get_current_user, require_csrf
from app.services.backup_service import BackupService

router = APIRouter(prefix="/backups", tags=["backups"], dependencies=[Depends(get_current_user)])


@router.post("", dependencies=[Depends(require_csrf)])
def create_backup() -> dict:
    try:
        return {"backup": BackupService().create_backup()}
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(exc)) from exc
