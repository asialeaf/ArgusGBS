<template>
<section class="content">
    <router-view @show-session-list="(filter, sms) => $emit('show-session-list', filter, sms)"
        @show-stream-stats="(title, serial, code) => $emit('show-stream-stats', title, serial, code)">
    </router-view>
</section>
</template>

<script>
export default {
    data() {
        return {}
    }
}
</script>
