<template>
    <button @click="router.push({ name: 'Dashboard', params: { title: component.title } })">
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