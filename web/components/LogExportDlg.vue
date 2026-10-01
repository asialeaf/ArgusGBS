<template>
    <FormDlg title="导出操作日志列表" @hide="onHide" @show="onShow" @submit="onSubmit" ref="dlg" :disabled="errors.any()">
        <div :class="{'form-group':true}">
            <div class="col-sm-12 checkbox text-center">
                <el-checkbox style="margin-left:-19px;margin-top:-5px;" v-model.trim="filter">
                    通过当前查询条件过滤后导出
                </el-checkbox>
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('start')}">
            <label for="input-start" class="col-sm-3 control-label">开始
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" id="input-start" name="start" v-model.trim="start" data-vv-as="开始" v-validate="'numeric|min_value:0'">
            </div>
        </div>
        <div :class="{'form-group':true,'has-error': errors.has('limit')}">
            <label for="input-limit" class="col-sm-3 control-label">上限
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" id="input-limit" name="limit" v-model.trim="limit" data-vv-as="上限" v-validate="'numeric|min_value:1'">
            </div>
        </div>
    </FormDlg>
</template>

<script>
import FormDlg from 'components/FormDlg.vue';
import $ from 'jquery';

export default {
    // props: {
    //     paging: {
    //         type: Boolean,
    //         default: false
    //     }
    // },
    data() {
        return {
            filter: true,
            start: 0,
            limit: 10000,
        }
    },
    components: { FormDlg },
    methods: {
        onHide() {
            this.filter = true;
            this.start = 0;
            this.limit = 10000;
            this.$emit("hide");
        },
        onShow() {
            this.errors.clear();
            this.$emit("show");
        },
        async onSubmit() {
            var ok = await this.$validator.validateAll();
            if(!ok) {
                var e = this.errors.items[0];
                this.$message({
                    type: 'error',
                    message: e.msg
                })
                $(`[name=${e.field}]`).focus();
                return;
            }
            let filter = this.filter;
            let start = this.start;
            let limit = this.limit;
            this.$refs['dlg'].hide();
            this.$emit("submit", filter, start, limit);
        },
        show() {
            this.errors.clear();
            this.$refs['dlg'].show();
        }
    }
}
</script>
