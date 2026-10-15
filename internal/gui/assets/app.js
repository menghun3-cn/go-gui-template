const out = document.querySelector("#out");

function setText(id, value) {
  const el = document.querySelector(id);
  if (el) el.textContent = value;
}

async function call(path, body) {
  const res = await fetch("/api" + path, {
    method: body ? "POST" : "GET",
    headers: body ? { "Content-Type": "application/json" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || res.statusText);
  return data;
}

async function loadState() {
  try {
    const state = await call("/state");
    setText("#version", state.version);
    setText("#pipe", state.namedPipe ? "on" : "off");
    setText("#server", "running");
  } catch (e) {
    setText("#server", "error");
    say(String(e));
  }
}

function say(text) {
  out.textContent = text;
}

document.querySelector("#checkUpdate").addEventListener("click", async () => {
  say("checking...");
  try {
    const data = await call("/update/check", {});
    say(data.upToDate ? "up to date" : "update found: " + data.version);
  } catch (e) {
    say(String(e));
  }
});

document.querySelector("#setLogin").addEventListener("click", async () => {
  try {
    const data = await call("/autostart", { enable: true });
    say("autostart " + (data.enabled ? "enabled" : "failed"));
  } catch (e) {
    say(String(e));
  }
});

loadState();
