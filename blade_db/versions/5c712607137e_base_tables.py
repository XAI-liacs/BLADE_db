"""Base Tables.

Revision ID: 5c712607137e
Revises: 
Create Date: 2026-05-29 11:45:22.935912

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql


# revision identifiers, used by Alembic.
revision: str = '5c712607137e'
down_revision: Union[str, Sequence[str], None] = None
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    """Experiment Table.
    CREATE TABLE Experiment(
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL UNIQUE,
        start_date TIMESTAMP NOT NULL,
        end_date TIMESTAMP NOT NULL
    );
    """
    op.create_table(
        "experiment",
        sa.Column('id', sa.Integer, primary_key=True),
        sa.Column('name', sa.Text(), nullable=False, unique=True),
        sa.Column('start_date', sa.TIMESTAMP(), nullable=False),
        sa.Column('end_date', sa.TIMESTAMP(), nullable=False)
    )

    """Run Table
    CREATE TABLE Run(
        id INTEGER PRIMARY KEY,
        seed INTEGER NOT NULL
    );
    """
    op.create_table(
        'run',
        sa.Column('id', sa.Integer, primary_key=True),
        sa.Column('seed', sa.Integer, nullable=False)
    )

    """Tag Table
    CREATE TABLE Tag(
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL
    );
    """
    op.create_table(
        'tag',
        sa.Column('id', sa.Integer, primary_key=True),
        sa.Column('name', sa.Text, nullable=False, unique=True)
    )

    """Message Table
    CREATE TABLE Message(
        id INTEGER PRIMARY KEY,
        message TEXT NOT NULL,
        created_at TIMESTAMP NOT NULL
    );
    """
    op.create_table(
        'message',
        sa.Column('id', sa.Integer, primary_key=True),
        sa.Column('message', sa.Text, nullable=False),
        sa.Column('created_at', sa.TIMESTAMP, nullable=False)
    )

    """Problem Table
    CREATE TABLE Problem(
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        prompt TEXT NOT NULL,
        evaluator TEXT NOT NULL,
        minimisation BOOLEAN NOT NULL,
        config JSONB
    );
"""
    op.create_table(
        'problem',
        sa.Column('id', sa.Integer, primary_key=True),
        sa.Column('name', sa.Text, nullable=False),
        sa.Column('prompt', sa.Text, nullable=False),
        sa.Column('evaluator', sa.Text, nullable=False),
        sa.Column('minimisation', sa.Boolean, nullable=False),
        sa.Column('config', postgresql.JSONB, nullable=True)
    )

    """Method Table
    CREATE TABLE Method(
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        source TEXT NOT NULL,
        config JSONB NOT NULL
    );
    """
    op.create_table(
        'method',
        sa.Column('id', sa.Integer, primary_key=True),
        sa.Column('name', sa.Text, nullable=False),
        sa.Column('source', sa.Text, nullable=False),
        sa.Column('config', postgresql.JSONB, nullable=False)
    )


    """LLM Table
    CREATE TABLE LLM(
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        hardware JSONB,
        config JSONB NOT NULL,
        UNIQUE (name, hardware, config)
    );
    """
    op.create_table(
        'llm',
        sa.Column('id', sa.Integer, primary_key=True),
        sa.Column('name', sa.Text, nullable=False),
        sa.Column('hardware', postgresql.JSONB, nullable=True),
        sa.Column('config', postgresql.JSONB, nullable=False),
        sa.UniqueConstraint('name', 'hardware', 'config')
    )

    """Solution Table
        CREATE TABLE Solution(
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            code TEXT NOT NULL,
            generation INTEGER,
            metadata JSONB,
            fitness JSONB NOT NULL,
            prompt TEXT
        );
    """
    op.create_table(
        "solution",
        sa.Column("id", sa.Integer, primary_key=True),
        sa.Column("name", sa.Text, nullable=False),
        sa.Column("description", sa.Text, nullable=True),
        sa.Column("code", sa.Text, nullable=False),
        sa.Column("generation", sa.Integer, nullable=True),
        sa.Column("metadata", postgresql.JSONB, nullable=True),
        sa.Column("fitness", postgresql.JSONB, nullable=False),
        sa.Column("prompt", sa.Text, nullable=True),
    )



def downgrade() -> None:
    """Downgrade schema."""
    op.drop_table("experiment")
    op.drop_table("run")
    op.drop_table("tag")
    op.drop_table("message")
    op.drop_table("problem")
    op.drop_table("method")
    op.drop_table("llm")
    op.drop_table("solution")
