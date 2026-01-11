<template>
    <button @click="navigateToDirectory" :class="component.cssClass">
        <span class="icon" v-html="unicodeIcon"></span>
        <span class="title">{{ component.title }}</span>
    </button>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { computed } from 'vue'

const router = useRouter()

const props = defineProps({
    component: {
        type: Object,
        required: true
    }
})

function getUnicodeIcon(icon) {
    if (icon === '' || !icon) {
        return '&#x1f4c1;' // Default folder icon
    } else {
        return unescape(icon)
    }
}

const unicodeIcon = computed(() => {
    return getUnicodeIcon(props.component.icon)
})

function navigateToDirectory() {
    const params = { title: props.component.title }
    
    if (props.component.entityType && props.component.entityKey) {
        params.entityType = props.component.entityType
        params.entityKey = props.component.entityKey
    }
    
    router.push({ name: 'Dashboard', params })
}
</script>

<style scoped>
.folder-container {
    display: grid;
}

button {
    display: flex;
    flex-direction: column;
    flex-grow: 1;
    justify-content: center;
    padding: 0.5em;
    box-shadow: 0 0 .6em #aaa;
    background-color: #fff;
    border-radius: .7em;
    border: 1px solid #ccc;
    cursor: pointer;
    transition: all 0.2s ease;
    font-size: .85em;
}

button:hover {
    background-color: #f5f5f5;
    border-color: #999;
}

button .icon {
    font-size: 3em;
    flex-grow: 1;
    align-content: center;
}

button .title {
    font-weight: 500;
    padding: 0.2em;
}

@media (prefers-color-scheme: dark) {
    button {
        box-shadow: 0 0 .6em #000;
        background-color: #111;
        border-color: #000;
        color: #fff;
    }

    button:hover {
        background-color: #222;
        border-color: #000;
        box-shadow: 0 0 6px #444;
        color: #fff;
    }
}

</style>