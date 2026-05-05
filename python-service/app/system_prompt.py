# Рабочий бэкап, генератор основа !!!
# Elfkbnm !!!
from langchain_core.prompts import ChatPromptTemplate

system_PROMPT = """
Ты извлекаешь триплеты для графа знаний.

Возвращай только JSON.

Формат:

[
 {"subject": "...", "relation": "...", "object": "..."}
]

Правила:

- избегай местоимений
- объединяй схожие сущности
- отношения должны быть глаголами или краткими фразами (НЕ БОЛЕЕ 4 СЛОВ)
- не выводи пояснения
"""

PROMPT = ChatPromptTemplate.from_messages(
    [
        ("system", system_PROMPT),
        ("human", "Text:\n{text}")
    ]
)