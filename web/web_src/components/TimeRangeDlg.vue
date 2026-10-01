<template>
    <div class="modal" ref="dlg" data-backdrop="static" data-disable="false" data-keyboard="true" tabindex="-1">
        <div :class="['modal-dialog', size]">
            <div class="modal-content">
                <div class="modal-header">
                    <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                        <span aria-hidden="true">&times;</span>
                    </button>
                    <h4 class="modal-title text-center text-primary"><span>{{title}}</span></h4>
                </div>
                <div class="modal-body">
                    <div class="text-center">
                        <el-time-picker is-range :arrow-control="arrow" v-if="visiable"
                            :value-format="format" :default-value="vdefault" v-model="value"
                            start-placeholder="开始时间" end-placeholder="结束时间">
                        </el-time-picker>
                    </div>
                </div>
                <div class="modal-footer">
                    <slot name="footer"></slot>
                    <button type="button" class="btn btn-default" data-dismiss="modal">取消</button>
                    <button type="button" class="btn btn-primary" @click.prevent="doSubmit" :disabled="!value">确定</button>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    import 'jquery-ui/ui/widgets/draggable'
    import $ from 'jquery'

    export default {
        props: {
            size: {
                type: String,
                default: ''
            },
            title: {
                type: String,
                default: '选择时间段'
            },
            arrow: {
                type: Boolean,
                default: false
            },
            format: {
                type: String,
                default: 'HH:mm:ss'
            },
        },
        data() {
            return {
                visiable: false,
                value: "",
                vdefault: undefined,
            }
        },
        mounted() {
            $(this.$el).find('.modal-content').draggable({
                handle: '.modal-header',
                cancel: ".modal-title span",
                addClasses: false,
                containment: 'document',
                delay: 100,
                opacity: 0.5
            });
            $(this.$el).on("shown.bs.modal", () => {
                this.visiable = true;
                this.$emit("show");
            }).on("hidden.bs.modal", () => {
                this.visiable = false;
                this.value = "";
                this.$emit("hide");
            })
        },
        methods: {
            show() {
                $(this.$el).modal("show");
            },
            hide() {
                $(this.$el).modal("hide");
            },
            setValue(v) {
                this.value = v;
            },
            setDefault(v) {
                this.vdefault = v;
            },
            doSubmit() {
                this.$emit('submit', this.value);
            }
        }
    }
</script>

<style lang="less" scoped>
    .modal-content {
        overflow: hidden;
    }

    @media screen and(min-width: 992px) {
        .modal-dialog.modal-lgg {
            width: 90%;
        }
    }

    @media screen and(min-width: 1200px) {
        .modal-dialog.modal-lgg {
            width: 1200px;
        }
    }

    @media screen and(max-width: 767px) {
        .el-date-editor--daterange.el-input, .el-date-editor--daterange.el-input__inner, .el-date-editor--timerange.el-input, .el-date-editor--timerange.el-input__inner {
            width: 250px;
        }
    }
</style>