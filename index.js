let currentConversationId = null;
let pending = false;

const conversationsEl = document.getElementById("conversations");
const messagesEl = document.getElementById("messages");
const promptEl = document.getElementById("prompt");
const sendEl = document.getElementById("send");
const formEl = document.getElementById("composer");

const loadConversations = async () => {
  const response = await fetch("/api/conversations");
  const list = await response.json();
  conversationsEl.innerHTML = "";
  for (const conversation of list) {
    const li = document.createElement("li");
    li.className =
      conversation.id === currentConversationId
        ? "conversation conversation--active"
        : "conversation";

    const title = document.createElement("span");
    title.className = "conversation__title";
    title.textContent = conversation.title || "Untitled";
    title.onclick = () => openConversation(conversation.id);

    const remove = document.createElement("button");
    remove.type = "button";
    remove.className = "conversation__delete";
    remove.textContent = "x";
    remove.title = "Delete";
    remove.onclick = (event) => {
      event.stopPropagation();
      deleteConversation(conversation.id);
    };

    li.appendChild(title);
    li.appendChild(remove);
    conversationsEl.appendChild(li);
  }
};

const addMessage = (modifiers, content, isHTML) => {
  const message = document.createElement("div");
  message.className = [
    "message",
    ...modifiers.split(" ").map((modifier) => `message--${modifier}`),
  ].join(" ");
  if (isHTML) {
    message.innerHTML = content;
  } else {
    message.textContent = content;
  }
  messagesEl.appendChild(message);
  messagesEl.scrollTop = messagesEl.scrollHeight;
  return message;
};

const newChat = () => {
  currentConversationId = null;
  messagesEl.innerHTML = "";
  promptEl.focus();
  loadConversations();
};

const openConversation = async (id) => {
  const response = await fetch(`/api/conversation?id=${id}`);
  if (!response.ok) {
    return;
  }
  const conversation = await response.json();
  currentConversationId = conversation.id;
  messagesEl.innerHTML = "";
  for (const exchange of conversation.exchanges) {
    addMessage("user", exchange.prompt, false);
    addMessage("assistant", exchange.html, true);
  }
  loadConversations();
};

const deleteConversation = async (id) => {
  if (!confirm("Delete this chat?")) {
    return;
  }
  await fetch("/api/conversation/delete", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ id: id }),
  });
  if (id === currentConversationId) {
    newChat();
  } else {
    loadConversations();
  }
};

const send = async (prompt) => {
  pending = true;
  sendEl.disabled = true;
  promptEl.disabled = true;

  addMessage("user", prompt, false);
  const thinking = addMessage("assistant thinking", "Thinking...", false);

  try {
    const response = await fetch("/api/ask", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        conversationId: currentConversationId || 0,
        prompt: prompt,
      }),
    });
    const result = await response.json();
    thinking.remove();
    if (result.error) {
      addMessage("assistant", result.error, false);
    } else {
      currentConversationId = result.conversationId;
      addMessage("assistant", result.html, true);
      loadConversations();
    }
  } catch (error) {
    thinking.remove();
    addMessage("assistant", `Request failed: ${error}`, false);
  } finally {
    pending = false;
    sendEl.disabled = false;
    promptEl.disabled = false;
    promptEl.focus();
  }
};

formEl.addEventListener("submit", (event) => {
  event.preventDefault();
  if (pending) {
    return;
  }
  const prompt = promptEl.value.trim();
  if (!prompt) {
    return;
  }
  promptEl.value = "";
  send(prompt);
});

promptEl.addEventListener("keydown", (event) => {
  if (event.key === "Enter" && !event.shiftKey) {
    event.preventDefault();
    formEl.requestSubmit();
  }
});

document.getElementById("new-chat").onclick = newChat;

loadConversations();
