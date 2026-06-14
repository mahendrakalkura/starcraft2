"use strict";

let currentConversationId = null;
let pending = false;

const conversationsEl = document.getElementById("conversations");
const messagesEl = document.getElementById("messages");
const promptEl = document.getElementById("prompt");
const sendEl = document.getElementById("send");
const formEl = document.getElementById("composer");

async function loadConversations() {
  const response = await fetch("/api/conversations");
  const list = await response.json();
  conversationsEl.innerHTML = "";
  for (const conversation of list) {
    const li = document.createElement("li");
    li.className = conversation.id === currentConversationId ? "active" : "";

    const title = document.createElement("span");
    title.className = "title";
    title.textContent = conversation.title || "Untitled";
    title.onclick = () => openConversation(conversation.id);

    const remove = document.createElement("button");
    remove.className = "delete";
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
}

function addMessage(role, content, isHTML) {
  const message = document.createElement("div");
  message.className = "message " + role;
  if (isHTML) {
    message.innerHTML = content;
  } else {
    message.textContent = content;
  }
  messagesEl.appendChild(message);
  messagesEl.scrollTop = messagesEl.scrollHeight;
  return message;
}

function newChat() {
  currentConversationId = null;
  messagesEl.innerHTML = "";
  promptEl.focus();
  loadConversations();
}

async function openConversation(id) {
  const response = await fetch("/api/conversation?id=" + id);
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
}

async function deleteConversation(id) {
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
}

async function send(prompt) {
  pending = true;
  sendEl.disabled = true;
  promptEl.disabled = true;

  addMessage("user", prompt, false);
  const thinking = addMessage("assistant thinking", "Thinking...", false);

  try {
    const response = await fetch("/api/ask", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ conversationId: currentConversationId || 0, prompt: prompt }),
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
    addMessage("assistant", "Request failed: " + error, false);
  } finally {
    pending = false;
    sendEl.disabled = false;
    promptEl.disabled = false;
    promptEl.focus();
  }
}

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
