<template>
    <section class = "transparent">
        <fieldset>
            <legend>{{ dashboard.title }}</legend>

            <ActionButton :actionData = "action" v-for = "action in dashboard.contents" :key = "action.id" />
        </fieldset>
    </section>
</template>

<script setup>
import ActionButton from './ActionButton.vue'

defineProps({
    dashboard: {
        type: Object,
        required: true
    }
})


</script>

<style>
fieldset {
	display: grid;
	grid-template-columns: repeat(auto-fit, 180px);
	grid-auto-rows: 1fr;
	justify-content: center;
	place-items: stretch;
}   
</style>