<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <label class="col-sm-4 control-label">SD卡编号
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-input-number size="small" v-model="id" :min="0"></el-input-number>
            </div>
        </div>
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-danger" @click.prevent="send" :disabled="sending">存储卡格式化<span v-show="sending">...</span></button>
                <button role="button" class="btn btn-info" @click.prevent="query" :disabled="loading">状态查询<span v-show="loading">...</span></button>
            </div>
        </div>
    </div>
</div>
</template>

<script>
export default {
    data() {
        return {
            sending: false,
            loading: false,
            id: 1,
        }
    },
    methods: {
        async send() {
            this.sending = true;
            await this.$store.dispatch("connect", {
                topic: "存储卡格式化",
            });

            $.post("/api/v1/control/formatsdcard", {
                serial: this.$store.state.serial,
                code: this.$store.state.code||this.$store.state.serial,
                id: this.id,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "存储卡格式化成功"
                })
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).always(() => {
                this.$store.commit("updateResult", null);
                this.sending = false;
            })
        },
        async query() {
            this.loading = true;
            await this.$store.dispatch("connect", {
                topic: "存储卡状态查询",
            });

            $.get("/api/v1/device/fetchsdcardstatus", {
                serial: this.$store.state.serial,
                code: this.$store.state.code||this.$store.state.serial,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "存储卡状态查询成功"
                })
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).always(() => {
                this.loading = false;
            })
        },
    },
}
</script>
