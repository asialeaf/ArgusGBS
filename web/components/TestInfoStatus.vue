<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-primary" @click.prevent="fetchstatus" :disabled="fetchstatusing">状态查询<span v-show="fetchstatusing">...</span></button>
                <button role="button" class="btn btn-info" @click.prevent="fetchinfo" :disabled="fetchinfoing">信息查询<span v-show="fetchinfoing">...</span></button>
            </div>
        </div>
    </div>
</div>
</template>

<script>

export default {
    data() {
        return {
            fetchinfoing: false,
            fetchstatusing: false,
        }
    },
    methods: {
        async fetchinfo() {
            this.fetchinfoing = true;
            await this.$store.dispatch("connect", {
                topic: "信息查询",
            });

            $.get("/api/v1/device/fetchinfo", {
                serial: this.$store.state.serial,
                code: this.$store.state.code||this.$store.state.serial,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "信息查询成功"
                })
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).always(() => {
                this.fetchinfoing = false;
            })
        },
        async fetchstatus() {
            this.fetchstatusing = true;
            await this.$store.dispatch("connect", {
                topic: "状态查询",
            });

            $.get("/api/v1/device/fetchstatus", {
                serial: this.$store.state.serial,
                code: this.$store.state.code||this.$store.state.serial,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "状态查询成功"
                })
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).always(() => {
                this.fetchstatusing = false;
            })
        }
    }
}
</script>
