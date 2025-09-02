<template>
    <div v-if="!dashboard" style = "text-align: center">
        <p>Loading... {{ title }}</p>
    </div>
    <div v-else>
        <section v-if="dashboard.contents.length == 0">
            <legend>{{ dashboard.title }}</legend>
            <p>This dashboard is empty.</p>
        </section>

        <section class="transparent" v-else>
            <div v-for="component in dashboard.contents" :key="component.title">
                <div v-if="component.type == 'fieldset'">
                    <fieldset>
                        <legend v-if = "dashboard.title != 'Default'">{{ component.title }}</legend>

                        <template v-for="subcomponent in component.contents">
                            <div v-if="subcomponent.type == 'display'" class="display">
                                <div v-html="subcomponent.title" />
                            </div>

                            <ActionButton v-else-if="subcomponent.type == 'link'" :actionData="subcomponent.action"
                                :key="subcomponent.title" />

                            <div v-else-if="subcomponent.type == 'directory'">
                                <router-link :to="{ name: 'Dashboard', params: { title: subcomponent.title } }"
                                    class="dashboard-link">
                                    <button>
                                        {{ subcomponent.title }}
                                    </button>
                                </router-link>
                            </div>

                            <div v-else>
                                OTHER: {{ subcomponent.type }}
                                {{ subcomponent }}
                            </div>
                        </template>
                    </fieldset>
                </div>

                <ActionButton v-else :actionData="action" v-for="action in component.contents" :key="action.title" />
            </div>
        </section>
    </div>
</template>

<script setup>
import ActionButton from './ActionButton.vue'
import { onMounted, ref } from 'vue'

const props = defineProps({
    title: {
        type: String,
        required: true
    }
})

const dashboard = ref(null)

async function getDashboard() {
    console.log("getting dashboard", props.title)
    const ret = await window.client.getDashboard({
        title: props.title,
    })

    dashboard.value = ret.dashboard
}

onMounted(() => {
    getDashboard()
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