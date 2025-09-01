document.addEventListener("DOMContentLoaded", function () {
  const form = document.querySelector(".notification-form");
  const textarea = document.querySelector(".apple-textarea");
  const telegramCheckbox = document.getElementById("telegram-checkbox");
  const emailCheckbox = document.getElementById("email-checkbox");
  const telegramInput = document.getElementById("telegram_id");
  const emailInput = document.getElementById("email");
  const telegramConfirmed = document.getElementById("telegram-confirmed");
  const confirmError = document.getElementById("confirm-error");
  const submitButton = document.querySelector(".apple-button");

  // Авто-высота textarea
  textarea.addEventListener("input", function () {
    this.style.height = "auto";
    this.style.height = this.scrollHeight + "px";
  });

  // Показать/скрыть поля Telegram
  telegramCheckbox.addEventListener("change", function () {
    const telegramFields = document.getElementById("telegram-fields");
    if (this.checked) {
      telegramFields.style.display = "block";
      telegramInput.setAttribute("required", "required");
    } else {
      telegramFields.style.display = "none";
      telegramInput.removeAttribute("required");
      telegramInput.value = "";
      telegramConfirmed.checked = false;
      confirmError.style.display = "none";
    }
    removeErrors();
  });

  // Показать/скрыть поля Email
  emailCheckbox.addEventListener("change", function () {
    const emailFields = document.getElementById("email-fields");
    if (this.checked) {
      emailFields.style.display = "block";
      emailInput.setAttribute("required", "required");
    } else {
      emailFields.style.display = "none";
      emailInput.removeAttribute("required");
      emailInput.value = "";
    }
    removeErrors();
  });

  // Валидация чекбокса подтверждения Telegram
  telegramConfirmed.addEventListener("change", function () {
    if (this.checked) {
      confirmError.style.display = "none";
    }
  });

  // Обработка отправки формы
  form.addEventListener("submit", async function (e) {
    e.preventDefault();

    const v = document.getElementById("notify_at").value;
    const message = textarea.value.trim();
    const isTelegramChecked = telegramCheckbox.checked;
    const isEmailChecked = emailCheckbox.checked;
    const telegramId = telegramInput.value.trim();
    const emailValue = emailInput.value.trim();

    // 2. Преобразуем локальное время в UTC
    const localDateTime = new Date(v);

    // Проверяем, что время корректно
    if (isNaN(localDateTime.getTime())) {
      alert("Некорректный формат времени");
      return;
    }
    const dt = localDateTime.toISOString();

    // Валидация
    if (
      !validateForm(
        message,
        isTelegramChecked,
        isEmailChecked,
        telegramId,
        emailValue,
        dt
      )
    ) {
      return;
    }

    setLoadingState(true);

    try {
      const response = await sendDataToServer({
        message,
        date: dt,
        telegram_set: isTelegramChecked,
        telegram_id: telegramId,
        telegram_confirmed: isTelegramChecked
          ? telegramConfirmed.checked
          : true, // всегда true если не используется
        email_set: isEmailChecked,
        email: emailValue,
      });

      showSuccess("Уведомление успешно создано!");
      form.reset();
      telegramConfirmed.checked = false;
    } catch (error) {
      handleRequestError(error);
    } finally {
      setLoadingState(false);
    }
  });

  function validateForm(
    message,
    isTelegramChecked,
    isEmailChecked,
    telegramId,
    emailValue,
    notifyAt
  ) {
    removeErrors();
    let isValid = true;

    if (!message) {
      showError(textarea, "Введите текст уведомления");
      isValid = false;
    }

    if (!notifyAt) {
      showError(document.getElementById("notify_at"), "Укажите дату и время");
      isValid = false;
    }

    if (!isTelegramChecked && !isEmailChecked) {
      showError(
        document.querySelector(".channels-container"),
        "Выберите хотя бы один способ отправки"
      );
      isValid = false;
    }

    if (isTelegramChecked && !telegramId) {
      showError(telegramInput, "Введите Telegram ID или @username");
      isValid = false;
    }

    if (isTelegramChecked && !telegramConfirmed.checked) {
      confirmError.style.display = "block";
      isValid = false;
    }

    if (isEmailChecked) {
      if (!emailValue) {
        showError(emailInput, "Введите email адрес");
        isValid = false;
      } else {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(emailValue)) {
          showError(emailInput, "Введите корректный email");
          isValid = false;
        }
      }
    }

    return isValid;
  }

  async function sendDataToServer(data) {
    const response = await fetch("/notify", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    });
    const rspClose = response.clone();
    console.log(rspClose.text);
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || `Ошибка: ${response.status}`);
    }

    return await response.json();
  }

  function handleRequestError(error) {
    console.error("Ошибка:", error);
    const msg = error.message.includes("Network")
      ? "Ошибка сети. Проверьте подключение."
      : error.message.includes("404")
      ? "Сервер не найден."
      : error.message.includes("500")
      ? "Ошибка сервера."
      : error.message.includes("400")
      ? "Неверно заполнены поля"
      : error.message || "ошибка отправки";

    showError(form, msg);
  }

  function showSuccess(message) {
    removeErrors();
    const successDiv = document.createElement("div");
    successDiv.className = "success-message";
    successDiv.textContent = message;
    successDiv.style.cssText = `
      color: #34C759;
      font-size: 16px;
      margin-top: 20px;
      padding: 15px;
      background: rgba(52, 199, 89, 0.1);
      border: 1px solid rgba(52, 199, 89, 0.3);
      border-radius: 12px;
    `;
    form.appendChild(successDiv);

    setTimeout(() => successDiv.remove(), 5000);
  }

  function showError(element, message) {
    const errorDiv = document.createElement("div");
    errorDiv.className = "error-message";
    errorDiv.textContent = message;
    errorDiv.style.cssText = `
      color: #FF3B30;
      font-size: 14px;
      margin-top: 8px;
    `;
    element.parentNode.insertBefore(errorDiv, element.nextSibling);
    element.style.animation = "shake 0.3s ease-in-out";
    setTimeout(() => (element.style.animation = ""), 300);
  }

  function removeErrors() {
    document
      .querySelectorAll(".error-message, .success-message")
      .forEach((el) => el.remove());
  }

  function setLoadingState(isLoading) {
    if (isLoading) {
      submitButton.disabled = true;
      submitButton.textContent = "Отправка...";
      submitButton.style.opacity = "0.7";
    } else {
      submitButton.disabled = false;
      submitButton.textContent = "Создать уведомление";
      submitButton.style.opacity = "1";
    }
  }

  textarea.focus();
});

// Анимации
const style = document.createElement("style");
style.textContent = `
  @keyframes shake {
    0%, 100% { transform: translateX(0); }
    25% { transform: translateX(-5px); }
    75% { transform: translateX(5px); }
  }
  .apple-button:disabled { cursor: not-allowed; }
`;
document.head.appendChild(style);
