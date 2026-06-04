"""HelpDesk 도메인 API: tickets + messages."""
from __future__ import annotations

from typing import Literal

from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel, Field, ConfigDict
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from .db import get_session
from .models import Message, Ticket, TicketPriority, TicketStatus
from .oauth import require_user_sub

router = APIRouter(prefix="/api")


# ── Schemas ──
class TicketCreate(BaseModel):
    subject: str = Field(..., min_length=1, max_length=200)
    priority: TicketPriority = TicketPriority.normal
    initial_message: str | None = None


class TicketUpdate(BaseModel):
    status: TicketStatus | None = None
    priority: TicketPriority | None = None


class MessageCreate(BaseModel):
    body: str = Field(..., min_length=1)


class MessageOut(BaseModel):
    model_config = ConfigDict(from_attributes=True)
    id: int
    author_sub: str
    body: str
    created_at: str

    @classmethod
    def from_model(cls, m: Message) -> "MessageOut":
        return cls(
            id=m.id,
            author_sub=m.author_sub,
            body=m.body,
            created_at=m.created_at.isoformat(),
        )


class TicketOut(BaseModel):
    id: int
    subject: str
    status: TicketStatus
    priority: TicketPriority
    created_at: str
    updated_at: str
    message_count: int

    @classmethod
    def from_model(cls, t: Ticket, message_count: int) -> "TicketOut":
        return cls(
            id=t.id,
            subject=t.subject,
            status=t.status,
            priority=t.priority,
            created_at=t.created_at.isoformat(),
            updated_at=t.updated_at.isoformat(),
            message_count=message_count,
        )


class TicketDetail(TicketOut):
    messages: list[MessageOut]


# ── Routes ──
@router.get("/tickets")
async def list_tickets(
    status: TicketStatus | None = None,
    priority: TicketPriority | None = None,
    sub: str = Depends(require_user_sub),
    session: AsyncSession = Depends(get_session),
):
    stmt = (
        select(Ticket)
        .where(Ticket.owner_sub == sub)
        .options(selectinload(Ticket.messages))
        .order_by(Ticket.updated_at.desc())
    )
    if status is not None:
        stmt = stmt.where(Ticket.status == status)
    if priority is not None:
        stmt = stmt.where(Ticket.priority == priority)

    items = list((await session.scalars(stmt)).all())
    return [TicketOut.from_model(t, len(t.messages)) for t in items]


@router.post("/tickets", status_code=201)
async def create_ticket(
    body: TicketCreate,
    sub: str = Depends(require_user_sub),
    session: AsyncSession = Depends(get_session),
):
    ticket = Ticket(owner_sub=sub, subject=body.subject.strip(), priority=body.priority)
    if body.initial_message:
        ticket.messages.append(Message(author_sub=sub, body=body.initial_message.strip()))
    session.add(ticket)
    await session.commit()
    await session.refresh(ticket, attribute_names=["messages"])
    return TicketOut.from_model(ticket, len(ticket.messages))


@router.get("/tickets/{ticket_id}")
async def get_ticket(
    ticket_id: int,
    sub: str = Depends(require_user_sub),
    session: AsyncSession = Depends(get_session),
):
    t = await session.scalar(
        select(Ticket)
        .where(Ticket.id == ticket_id, Ticket.owner_sub == sub)
        .options(selectinload(Ticket.messages))
    )
    if t is None:
        raise HTTPException(status_code=404)
    base = TicketOut.from_model(t, len(t.messages))
    return TicketDetail(**base.model_dump(), messages=[MessageOut.from_model(m) for m in t.messages])


@router.patch("/tickets/{ticket_id}")
async def update_ticket(
    ticket_id: int,
    body: TicketUpdate,
    sub: str = Depends(require_user_sub),
    session: AsyncSession = Depends(get_session),
):
    t = await session.scalar(
        select(Ticket).where(Ticket.id == ticket_id, Ticket.owner_sub == sub)
    )
    if t is None:
        raise HTTPException(status_code=404)
    if body.status is not None:
        t.status = body.status
    if body.priority is not None:
        t.priority = body.priority
    await session.commit()
    await session.refresh(t, attribute_names=["messages"])
    return TicketOut.from_model(t, len(t.messages))


@router.delete("/tickets/{ticket_id}", status_code=204)
async def delete_ticket(
    ticket_id: int,
    sub: str = Depends(require_user_sub),
    session: AsyncSession = Depends(get_session),
):
    t = await session.scalar(
        select(Ticket).where(Ticket.id == ticket_id, Ticket.owner_sub == sub)
    )
    if t is None:
        raise HTTPException(status_code=404)
    await session.delete(t)
    await session.commit()


@router.post("/tickets/{ticket_id}/messages", status_code=201)
async def add_message(
    ticket_id: int,
    body: MessageCreate,
    sub: str = Depends(require_user_sub),
    session: AsyncSession = Depends(get_session),
):
    t = await session.scalar(
        select(Ticket).where(Ticket.id == ticket_id, Ticket.owner_sub == sub)
    )
    if t is None:
        raise HTTPException(status_code=404)
    msg = Message(ticket_id=t.id, author_sub=sub, body=body.body.strip())
    session.add(msg)
    # 메시지 추가시 ticket 갱신 시간을 자연스럽게 변경
    t.subject = t.subject
    await session.commit()
    await session.refresh(msg)
    return MessageOut.from_model(msg)
