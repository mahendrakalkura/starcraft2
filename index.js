const conversations = document.getElementById("conversations");
const messages = document.getElementById("messages");
const promptInput = document.getElementById("prompt");
const send = document.getElementById("send");
const form = document.getElementById("composer");

let currentConversationId = null;
let pending = false;

const loadConversations = async () => {
  const response = await fetch("/api/conversations");
  const list = await response.json();

  conversations.innerHTML = "";
  for (const conversation of list) {
    conversations.append(renderConversation(conversation));
  }
};

const renderConversation = (conversation) => {
  const item = document.createElement("li");
  const active = conversation.id === currentConversationId;
  item.className = active ? "conversation conversation--active" : "conversation";

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

  item.append(title, remove);
  return item;
};

const addMessage = (role, content, asHTML) => {
  const message = document.createElement("div");
  message.className = `message message--${role}`;
  if (asHTML) {
    message.innerHTML = content;
  } else {
    message.textContent = content;
  }

  messages.append(message);
  messages.scrollTop = messages.scrollHeight;
  return message;
};

const newChat = () => {
  currentConversationId = null;
  messages.innerHTML = "";
  promptInput.focus();
  loadConversations();
};

const openConversation = async (id) => {
  const response = await fetch(`/api/conversation?id=${id}`);
  if (!response.ok) {
    return;
  }

  const conversation = await response.json();
  currentConversationId = conversation.id;
  messages.innerHTML = "";
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
    body: JSON.stringify({ id }),
  });

  if (id === currentConversationId) {
    newChat();
  } else {
    loadConversations();
  }
};

const ask = async (prompt) => {
  pending = true;
  send.disabled = true;
  promptInput.disabled = true;
  addMessage("user", prompt, false);
  const thinking = addMessage("thinking", "Thinking...", false);

  try {
    const response = await fetch("/api/ask", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ conversationId: currentConversationId || 0, prompt }),
    });
    const result = await response.json();
    thinking.remove();

    if (result.error) {
      addMessage("assistant", result.error, false);
      return;
    }
    currentConversationId = result.conversationId;
    addMessage("assistant", result.html, true);
    loadConversations();
  } catch (error) {
    thinking.remove();
    addMessage("assistant", `Request failed: ${error}`, false);
  } finally {
    pending = false;
    send.disabled = false;
    promptInput.disabled = false;
    promptInput.focus();
  }
};

form.addEventListener("submit", (event) => {
  event.preventDefault();
  if (pending) {
    return;
  }

  const prompt = promptInput.value.trim();
  if (!prompt) {
    return;
  }

  promptInput.value = "";
  ask(prompt);
});

promptInput.addEventListener("keydown", (event) => {
  if (event.key === "Enter" && !event.shiftKey) {
    event.preventDefault();
    form.requestSubmit();
  }
});

document.getElementById("new-chat").onclick = newChat;

loadConversations();
