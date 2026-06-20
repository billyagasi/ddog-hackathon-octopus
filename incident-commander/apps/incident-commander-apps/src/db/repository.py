from sqlalchemy import create_engine, Column, Integer, String, Text, JSON, text
from sqlalchemy.orm import declarative_base, sessionmaker
from pgvector.sqlalchemy import Vector
from src.core.config import settings
import logging

logger = logging.getLogger(__name__)

Base = declarative_base()

class IncidentRecord(Base):
    __tablename__ = 'incidents'

    id = Column(Integer, primary_key=True)
    incident_id = Column(String, unique=True, index=True)
    service_name = Column(String)
    status = Column(String)
    timeline = Column(JSON)
    rca_document = Column(Text)
    rca_embedding = Column(Vector(1536))  # e.g., Amazon Titan embeddings dimension

def get_engine():
    if not settings.database_url:
        logger.warning("No DATABASE_URL configured.")
        return None
    return create_engine(settings.database_url)

def init_db():
    engine = get_engine()
    if engine:
        # Note: 'CREATE EXTENSION IF NOT EXISTS vector' should ideally be run manually or via migrations
        try:
            with engine.connect() as conn:
                conn.execute(text("CREATE EXTENSION IF NOT EXISTS vector"))
                conn.commit()
        except Exception as e:
            logger.error(f"Could not create vector extension: {e}")
        Base.metadata.create_all(engine)

def save_incident(incident_data: dict, rca_doc: str, embedding: list[float] = None):
    engine = get_engine()
    if not engine:
        return
    
    Session = sessionmaker(bind=engine)
    with Session() as session:
        record = IncidentRecord(
            incident_id=incident_data.get('incident_id'),
            service_name=incident_data.get('service_name'),
            status=incident_data.get('status'),
            timeline=incident_data.get('timeline'),
            rca_document=rca_doc,
            rca_embedding=embedding
        )
        session.add(record)
        session.commit()
        logger.info(f"Saved incident {record.incident_id} to database.")
