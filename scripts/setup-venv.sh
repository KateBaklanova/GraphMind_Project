#!/bin/bash

# Скрипт для создания venv и установки зависимостей

set -e  # Остановка при ошибке

echo "Setting up Python virtual environment..."

# Переходим в папку python-service
cd python-service

# Проверяем Python версию
python_version=$(python3 --version 2>&1 | grep -Po '(?<=Python )\d+\.\d+')
echo "Python version: $python_version"

# Создаем venv если не существует
if [ ! -d "venv" ]; then
    echo "Creating virtual environment..."
    python3 -m venv venv
else
    echo "Virtual environment already exists"
fi

# Активируем venv и устанавливаем зависимости
echo "Activating venv and installing dependencies..."
source venv/bin/activate

# Обновляем pip
pip install --upgrade pip

# Устанавливаем зависимости
pip install -r requirements.txt

# Устанавливаем дополнительные dev зависимости
if [ -f "requirements-dev.txt" ]; then
    pip install -r requirements-dev.txt
fi

echo "Virtual environment setup complete!"
echo ""
echo "To activate venv manually:"
echo "  cd python-service"
echo "  source venv/bin/activate"
echo ""
echo "To deactivate:"
echo "  deactivate"