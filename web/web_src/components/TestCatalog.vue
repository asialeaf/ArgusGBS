<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-primary" @click.prevent="query" :disabled="loading">发送<span v-show="loading">...</span></button>
            </div>
        </div>
    </div>
</div>
</template>

<script>

export default {
    data() {
        return {
            loading: false,
        }
    },
    methods: {
        async query() {
            this.loading = true;
            await this.$store.dispatch("connect");

            $.get("/api/v1/device/fetchcatalog", {
                serial: this.$store.state.serial,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "目录查询成功"
                })
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).always(() => {
                this.loading = false;
            })
        }
    }
}
</script>
