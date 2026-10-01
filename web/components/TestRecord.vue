<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <label class="col-sm-4 control-label">录像命令
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <select class="form-control" name="command" v-model.trim="command">
                    <option value="start">开始录像(Record)</option>
                    <option value="stop">停止录像(StopRecord)</option>
                </select>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">码流类型
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <select class="form-control" name="streamnumber" v-model.trim="streamnumber">
                    <option value="0">0-主码流</option>
                    <option value="1">1-子码流</option>
                </select>
            </div>
        </div>
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-primary" @click.prevent="send" :disabled="sending">发送<span v-show="sending">...</span></button>
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
            command: "start",
            streamnumber: 0,
        }
    },
    methods: {
        async send() {
            this.sending = true;
            await this.$store.dispatch("connect");

            $.post("/api/v1/control/record", {
                serial: this.$store.state.serial,
                code: this.$store.state.code,
                command: this.command,
                streamnumber: this.streamnumber,
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: "手动录像操作成功"
                })
                this.$store.commit("updateResult", ret);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).fail(ret => {
                this.$store.commit("updateResult", null);
            }).always(() => {
                this.sending = false;
            })
        },
    },
}
</script>