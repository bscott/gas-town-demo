// SlackLite Frontend

(function() {
    'use strict';

    // State
    const state = {
        username: null,
        channels: [],
        currentChannel: null,
        messages: [],
        ws: null
    };

    // DOM Elements
    const elements = {
        channelList: document.getElementById('channel-list'),
        currentChannelName: document.getElementById('current-channel-name'),
        messages: document.getElementById('messages'),
        messageForm: document.getElementById('message-form'),
        messageInput: document.getElementById('message-input'),
        addChannelBtn: document.getElementById('add-channel-btn'),
        modalOverlay: document.getElementById('modal-overlay'),
        channelForm: document.getElementById('channel-form'),
        channelNameInput: document.getElementById('channel-name-input'),
        cancelModalBtn: document.getElementById('cancel-modal-btn')
    };

    // Generate anonymous username
    function generateUsername() {
        const adjectives = ['happy', 'clever', 'swift', 'bright', 'calm', 'eager', 'bold', 'cool'];
        const nouns = ['panda', 'tiger', 'eagle', 'dolphin', 'fox', 'owl', 'wolf', 'bear'];
        const adj = adjectives[Math.floor(Math.random() * adjectives.length)];
        const noun = nouns[Math.floor(Math.random() * nouns.length)];
        const num = Math.floor(Math.random() * 100);
        return `${adj}-${noun}-${num}`;
    }

    // API helpers
    const api = {
        baseUrl: '/api',

        async getChannels() {
            const res = await fetch(`${this.baseUrl}/channels`);
            if (!res.ok) throw new Error('Failed to fetch channels');
            return res.json();
        },

        async createChannel(name) {
            const res = await fetch(`${this.baseUrl}/channels`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name })
            });
            if (!res.ok) throw new Error('Failed to create channel');
            return res.json();
        },

        async getMessages(channelId) {
            const res = await fetch(`${this.baseUrl}/channels/${channelId}/messages`);
            if (!res.ok) throw new Error('Failed to fetch messages');
            return res.json();
        },

        async sendMessage(channelId, content) {
            const res = await fetch(`${this.baseUrl}/channels/${channelId}/messages`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    username: state.username,
                    content
                })
            });
            if (!res.ok) throw new Error('Failed to send message');
            return res.json();
        }
    };

    // WebSocket connection
    function connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;

        state.ws = new WebSocket(wsUrl);

        state.ws.onopen = () => {
            console.log('WebSocket connected');
        };

        state.ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                handleWebSocketMessage(data);
            } catch (e) {
                console.error('Failed to parse WebSocket message:', e);
            }
        };

        state.ws.onclose = () => {
            console.log('WebSocket disconnected, reconnecting in 3s...');
            setTimeout(connectWebSocket, 3000);
        };

        state.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    function handleWebSocketMessage(data) {
        switch (data.type) {
            case 'message':
                if (data.channel_id === state.currentChannel?.id) {
                    state.messages.push(data.message);
                    renderMessages();
                    scrollToBottom();
                }
                break;
            case 'channel_created':
                state.channels.push(data.channel);
                renderChannels();
                break;
            default:
                console.log('Unknown WebSocket message type:', data.type);
        }
    }

    // Safe DOM manipulation helpers
    function createChannelItem(channel, isActive) {
        const li = document.createElement('li');
        li.className = 'channel-item' + (isActive ? ' active' : '');
        li.dataset.id = channel.id;
        li.textContent = channel.name;
        return li;
    }

    function createMessageElement(msg) {
        const div = document.createElement('div');
        div.className = 'message';

        const avatar = document.createElement('div');
        avatar.className = 'message-avatar';
        avatar.textContent = msg.username.slice(0, 2).toUpperCase();

        const content = document.createElement('div');
        content.className = 'message-content';

        const header = document.createElement('div');
        header.className = 'message-header';

        const author = document.createElement('span');
        author.className = 'message-author';
        author.textContent = msg.username;

        const time = document.createElement('span');
        time.className = 'message-time';
        time.textContent = formatTime(msg.created_at);

        header.appendChild(author);
        header.appendChild(time);

        const text = document.createElement('div');
        text.className = 'message-text';
        text.textContent = msg.content;

        content.appendChild(header);
        content.appendChild(text);

        div.appendChild(avatar);
        div.appendChild(content);

        return div;
    }

    function createEmptyState(channelName) {
        const div = document.createElement('div');
        div.className = 'empty-state';

        const h3 = document.createElement('h3');
        h3.textContent = 'No messages yet';

        const p = document.createElement('p');
        p.textContent = `Be the first to send a message in #${channelName}`;

        div.appendChild(h3);
        div.appendChild(p);

        return div;
    }

    // Rendering
    function renderChannels() {
        elements.channelList.replaceChildren();
        state.channels.forEach(channel => {
            const item = createChannelItem(channel, channel.id === state.currentChannel?.id);
            elements.channelList.appendChild(item);
        });
    }

    function renderMessages() {
        elements.messages.replaceChildren();

        if (state.messages.length === 0) {
            elements.messages.appendChild(
                createEmptyState(state.currentChannel?.name || 'general')
            );
            return;
        }

        state.messages.forEach(msg => {
            elements.messages.appendChild(createMessageElement(msg));
        });
    }

    // Helpers
    function formatTime(timestamp) {
        const date = new Date(timestamp);
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }

    function scrollToBottom() {
        elements.messages.scrollTop = elements.messages.scrollHeight;
    }

    function showModal() {
        elements.modalOverlay.classList.remove('hidden');
        elements.channelNameInput.value = '';
        elements.channelNameInput.focus();
    }

    function hideModal() {
        elements.modalOverlay.classList.add('hidden');
    }

    // Event handlers
    async function handleChannelClick(e) {
        const item = e.target.closest('.channel-item');
        if (!item) return;

        const channelId = item.dataset.id;
        const channel = state.channels.find(c => c.id === channelId);
        if (!channel || channel.id === state.currentChannel?.id) return;

        state.currentChannel = channel;
        elements.currentChannelName.textContent = `#${channel.name}`;
        renderChannels();

        try {
            state.messages = await api.getMessages(channel.id);
            renderMessages();
            scrollToBottom();
        } catch (e) {
            console.error('Failed to load messages:', e);
        }
    }

    async function handleMessageSubmit(e) {
        e.preventDefault();
        const content = elements.messageInput.value.trim();
        if (!content || !state.currentChannel) return;

        elements.messageInput.value = '';

        try {
            await api.sendMessage(state.currentChannel.id, content);
        } catch (e) {
            console.error('Failed to send message:', e);
            elements.messageInput.value = content;
        }
    }

    async function handleChannelCreate(e) {
        e.preventDefault();
        const name = elements.channelNameInput.value.trim().toLowerCase().replace(/\s+/g, '-');
        if (!name) return;

        try {
            const channel = await api.createChannel(name);
            state.channels.push(channel);
            renderChannels();
            hideModal();

            // Switch to the new channel
            state.currentChannel = channel;
            elements.currentChannelName.textContent = `#${channel.name}`;
            state.messages = [];
            renderChannels();
            renderMessages();
        } catch (e) {
            console.error('Failed to create channel:', e);
            alert('Failed to create channel');
        }
    }

    // Initialize
    async function init() {
        // Generate username
        state.username = localStorage.getItem('slacklite_username');
        if (!state.username) {
            state.username = generateUsername();
            localStorage.setItem('slacklite_username', state.username);
        }

        // Set up event listeners
        elements.channelList.addEventListener('click', handleChannelClick);
        elements.messageForm.addEventListener('submit', handleMessageSubmit);
        elements.addChannelBtn.addEventListener('click', showModal);
        elements.cancelModalBtn.addEventListener('click', hideModal);
        elements.channelForm.addEventListener('submit', handleChannelCreate);
        elements.modalOverlay.addEventListener('click', (e) => {
            if (e.target === elements.modalOverlay) hideModal();
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') hideModal();
        });

        // Load channels
        try {
            state.channels = await api.getChannels();
            renderChannels();

            // Select first channel or create default
            if (state.channels.length > 0) {
                state.currentChannel = state.channels[0];
                elements.currentChannelName.textContent = `#${state.currentChannel.name}`;
                renderChannels();

                state.messages = await api.getMessages(state.currentChannel.id);
                renderMessages();
                scrollToBottom();
            }
        } catch (e) {
            console.error('Failed to load channels:', e);
        }

        // Connect WebSocket
        connectWebSocket();

        // Focus message input
        elements.messageInput.focus();
    }

    // Start app when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
