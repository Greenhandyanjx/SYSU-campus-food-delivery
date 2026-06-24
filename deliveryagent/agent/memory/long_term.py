"""
Memory: 长期记忆
持久化跨会话的记忆，使用 FAISS 进行语义检索（替代原 ChromaDB）
"""
import json
import os
from typing import Optional
from datetime import datetime

from langchain_core.documents import Document
from langchain_community.vectorstores import FAISS

from utils.logger import logger
from utils.config import chroma_conf
from utils.path_tool import get_abs_path
from model.factory import embed_model


class LongTermMemory:
    """
    长期记忆：跨会话持久化
    保存用户偏好、历史行为、重要结论
    通过语义相似度检索相关记忆
    """

    def __init__(self, collection_name: str = "long_term_memory"):
        self.collection_name = collection_name
        raw_dir = chroma_conf.get("persist_directory", "db/faiss_index")
        abs_dir = os.path.normpath(get_abs_path(raw_dir))
        os.makedirs(abs_dir, exist_ok=True)
        # FAISS C++ fopen 无法处理含中文的绝对路径（Windows），
        # 使用 relpath 确保传递给 FAISS 的是 ASCII 路径。
        self.persist_dir = os.path.join(os.path.relpath(abs_dir), collection_name)
        os.makedirs(self.persist_dir, exist_ok=True)

        self.vector_store: Optional[FAISS] = None
        try:
            index_file = os.path.join(self.persist_dir, "index.faiss")
            if os.path.exists(index_file):
                self.vector_store = FAISS.load_local(
                    self.persist_dir,
                    embed_model,
                    allow_dangerous_deserialization=True,
                )
                logger.info(f"[LongTermMemory] 加载 FAISS 索引: {self.persist_dir}")
            else:
                logger.info(f"[LongTermMemory] 初始化空存储: {collection_name}")
        except Exception as e:
            logger.warning(f"[LongTermMemory] 初始化失败（可能是首次使用）: {e}")
            self.vector_store = None

    def _save(self):
        """持久化 FAISS 索引到磁盘"""
        if self.vector_store is not None:
            self.vector_store.save_local(self.persist_dir)

    def remember(self, key: str, content: str, metadata: Optional[dict] = None) -> bool:
        """
        记住一条信息
        key: 记忆的唯一标识
        content: 记忆内容
        metadata: 附加元数据（如时间、来源等）
        """
        try:
            meta = metadata or {}
            meta["key"] = key
            meta["timestamp"] = datetime.now().isoformat()

            doc = Document(page_content=content, metadata=meta)
            if self.vector_store is None:
                self.vector_store = FAISS.from_documents([doc], embed_model)
            else:
                self.vector_store.add_documents([doc])
            self._save()
            logger.info(f"[LongTermMemory] 记住: {key}")
            return True
        except Exception as e:
            logger.error(f"[LongTermMemory] 保存失败: {e}")
            return False

    def recall(self, query: str, k: int = 3) -> list[dict]:
        """
        根据语义检索相关记忆
        """
        if self.vector_store is None:
            return []

        try:
            retriever = self.vector_store.as_retriever(search_kwargs={"k": k})
            docs = retriever.invoke(query)
            return [
                {
                    "content": doc.page_content,
                    "metadata": doc.metadata,
                    "score": getattr(doc, "score", None),
                }
                for doc in docs
            ]
        except Exception as e:
            logger.error(f"[LongTermMemory] 检索失败: {e}")
            return []

    def forget(self, key: str) -> bool:
        """
        忘记一条信息（FAISS 不支持单条删除，重新构建索引）
        """
        if self.vector_store is None:
            return False
        try:
            all_docs = list(self.vector_store.docstore._dict.values())
            kept_docs = [doc for doc in all_docs if doc.metadata.get("key") != key]

            if len(kept_docs) == len(all_docs):
                logger.info(f"[LongTermMemory] 未找到 key={key}，无需删除")
                return False

            if kept_docs:
                self.vector_store = FAISS.from_documents(kept_docs, embed_model)
            else:
                self.vector_store = None
                # 清空持久化文件
                for f in os.listdir(self.persist_dir):
                    os.remove(os.path.join(self.persist_dir, f))

            self._save() if kept_docs else None
            logger.info(f"[LongTermMemory] 遗忘: {key}")
            return True
        except Exception as e:
            logger.error(f"[LongTermMemory] 删除失败: {e}")
            return False

    def get_stats(self) -> dict:
        """获取长期记忆统计"""
        if self.vector_store is None:
            return {"status": "uninitialized", "count": 0}
        try:
            count = self.vector_store.index.ntotal
            return {"status": "ready", "count": count}
        except Exception as e:
            logger.error(f"[LongTermMemory] 获取统计失败: {e}")
            return {"status": "error", "count": 0}
