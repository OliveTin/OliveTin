<template>
    <div id = "breadcrumbs">
        <template v-for="(link, index) in links" :key="link.name">
            <router-link :to="link.href">{{ link.name }}</router-link>
            <span v-if="index < links.length - 1" class="separator">
                &raquo;
            </span>
        </template>
    </div>
</template>

<style scoped>
span {
    color: #bbb;
}

a {
    text-decoration: none;
    padding: 0.4em;
    border-radius: 0.2em;
}

a:hover {
    text-decoration: underline;
    background-color: #000;
}

</style>


<script setup>
    import { ref } from 'vue';
    import { watch } from 'vue';
    import { useRoute } from 'vue-router';

    const route = useRoute();
    const links = ref([]);

    watch(() => route.matched, (matched) => {

        links.value = [];
        matched.forEach((record) => {
            if (record.meta && record.meta.breadcrumb) {
                record.meta.breadcrumb.forEach((item) => {
                    links.value.push({
                        name: item.name,
                        href: item.href || record.path || '/'
                    });
                });
            } else if (record.name) {
                links.value.push({
                    name: record.name,
                    href: record.path || '/'
                });
            }
        });
    }, { immediate: true });
</script>
