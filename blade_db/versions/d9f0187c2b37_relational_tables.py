"""Relational Tables.

Revision ID: d9f0187c2b37
Revises: 5c712607137e
Create Date: 2026-05-29 13:36:22.038095

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
# from sqlalchemy.dialects import postgresql

# revision identifiers, used by Alembic.
revision: str = 'd9f0187c2b37'
down_revision: Union[str, Sequence[str], None] = '5c712607137e'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:

    """
    Table ExperimentRun
    CREATE TABLE ExperimentRun(
        experiment_id Integer NOT NULL 
            REFERENCES Experiment(id),
        run_id Integer NOT NULL 
            REFERENCES Run(id),
        UNIQUE (experiment_id, run_id)
    );
    """
    op.create_table(
        "experiment_run",
        sa.Column(
            "experiment_id",
            sa.Integer,
            sa.ForeignKey("experiment.id"),
            nullable=False,
        ),
        sa.Column(
            "run_id",
            sa.Integer,
            sa.ForeignKey("run.id"),
            nullable=False,
        ),
        sa.UniqueConstraint("experiment_id", "run_id"),
    )

    """
    Table ConversationLog
    CREATE TABLE ConversationLog(
        run_id Integer NOT NULL 
            REFERENCES Run(id),
        message_id Integer NOT NULL 
            REFERENCES Message(id),
        UNIQUE (run_id, message_id)
    );
    """
    op.create_table(
        "conversation_log",
        sa.Column(
            "run_id",
            sa.Integer,
            sa.ForeignKey("run.id"),
            nullable=False,
        ),
        sa.Column(
            "message_id",
            sa.Integer,
            sa.ForeignKey("message.id"),
            nullable=False,
        ),
        sa.UniqueConstraint("run_id", "message_id"),
    )

    """
    Table Sender
    CREATE TABLE Sender (
        message_id Integer NOT NULL
            REFERENCES Message(id),
        llm_id Integer
            REFERENCES LLM(id),
        method_id Integer
            REFERENCES Method(id),
        UNIQUE (message_id, llm_id, method_id),
        CHECK (
            (llm_id IS NOT NULL AND method_id IS NULL)
            OR
            (llm_id IS NULL AND method_id IS NOT NULL)
        )
    );
    """
    op.create_table(
        "sender",
        sa.Column(
            "message_id",
            sa.Integer,
            sa.ForeignKey("message.id"),
            nullable=False,
        ),
        sa.Column(
            "llm_id",
            sa.Integer,
            sa.ForeignKey("llm.id"),
            nullable=True,
        ),
        sa.Column(
            "method_id",
            sa.Integer,
            sa.ForeignKey("method.id"),
            nullable=True,
        ),
        sa.UniqueConstraint("message_id", "llm_id", "method_id"),
        sa.CheckConstraint(
            "(llm_id IS NOT NULL AND method_id IS NULL) OR "
            "(llm_id IS NULL AND method_id IS NOT NULL)",
        ),
    )

    """
    Table ProblemTags
    CREATE TABLE ProblemTags(
        tag_id Integer NOT NULL
            REFERENCES Tag(id),
        problem_id Integer NOT NULL
            REFERENCES Problem(id),
        UNIQUE (tag_id, problem_id)
    );
    """
    op.create_table(
        "problem_tags",
        sa.Column(
            "tag_id",
            sa.Integer,
            sa.ForeignKey("tag.id"),
            nullable=False,
        ),
        sa.Column(
            "problem_id",
            sa.Integer,
            sa.ForeignKey("problem.id"),
            nullable=False,
        ),
        sa.UniqueConstraint("tag_id", "problem_id"),
    )

    """
    Table MethodLLM
    CREATE TABLE MethodLLM(
        llm_id Integer NOT NULL
            REFERENCES LLM(id),
        method_id Integer NOT NULL
            REFERENCES Method(id),
        UNIQUE (llm_id, method_id)
    );
    """
    op.create_table(
        "method_llm",
        sa.Column(
            "llm_id",
            sa.Integer,
            sa.ForeignKey("llm.id"),
            nullable=False,
        ),
        sa.Column(
            "method_id",
            sa.Integer,
            sa.ForeignKey("method.id"),
            nullable=False,
        ),
        sa.UniqueConstraint("llm_id", "method_id"),
    )

    """
    Table RunDescriptor
    CREATE TABLE RunDescriptor(
        run_id Integer NOT NULL
            REFERENCES Run(id),
        method_id Integer NOT NULL
            REFERENCES Method(id),
        problem_id Integer NOT NULL
            REFERENCES Problem(id),
        UNIQUE (run_id, method_id, problem_id)
    );
    """
    op.create_table(
        "run_descriptor",
        sa.Column(
            "run_id",
            sa.Integer,
            sa.ForeignKey("run.id"),
            nullable=False,
        ),
        sa.Column(
            "method_id",
            sa.Integer,
            sa.ForeignKey("method.id"),
            nullable=False,
        ),
        sa.Column(
            "problem_id",
            sa.Integer,
            sa.ForeignKey("problem.id"),
            nullable=False,
        ),
        sa.UniqueConstraint("run_id", "method_id", "problem_id"),
    )

    """
    Table RunSolution
    CREATE TABLE RunSolution(
        run_id Integer NOT NULL
            REFERENCES Run(id),
        solution_id Integer NOT NULL
            REFERENCES Solution(id),
        UNIQUE (run_id, solution_id)
    );
    """
    op.create_table(
        "run_solution",
        sa.Column(
            "run_id",
            sa.Integer,
            sa.ForeignKey("run.id"),
            nullable=False,
        ),
        sa.Column(
            "solution_id",
            sa.Integer,
            sa.ForeignKey("solution.id"),
            nullable=False,
        ),
        sa.UniqueConstraint("run_id", "solution_id"),
    )

    """
    Table ParentChild
    CREATE TABLE ParentChild(
        parent_id Integer NOT NULL
            REFERENCES Solution(id),
        child_id Integer NOT NULL
            REFERENCES Solution(id),
        UNIQUE (parent_id, child_id)
    );
    """
    op.create_table(
        "parent_child",
        sa.Column(
            "parent_id",
            sa.Integer,
            sa.ForeignKey("solution.id"),
            nullable=False,
        ),
        sa.Column(
            "child_id",
            sa.Integer,
            sa.ForeignKey("solution.id"),
            nullable=False,
        ),
        sa.UniqueConstraint("parent_id", "child_id"),
    )


def downgrade() -> None:
    op.drop_table("parent_child")
    op.drop_table("run_solution")
    op.drop_table("run_descriptor")
    op.drop_table("method_llm")
    op.drop_table("problem_tags")
    op.drop_table("sender")
    op.drop_table("conversation_log")
    op.drop_table("experiment_run")
