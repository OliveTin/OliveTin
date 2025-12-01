<template>
    <button @click="navigateToDirectory">
        {{ component.title }}
    </button>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const props = defineProps({
    component: {
        type: Object,
        required: true
    }
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
    box-shadow: 0 0 .6em #aaa;
    background-color: #fff;
    border-radius: .7em;
}

button:hover {
    background-color: #f5f5f5;
    border-color: #999;
}

@media (prefers-color-scheme: dark) {
    button {
        box-shadow: 0 0 .6em #000;
        background-color: #111;
        border-color: #000;
    }

    button:hover {
        background-color: #222;
        border-color: #000;
    }
}

</style>