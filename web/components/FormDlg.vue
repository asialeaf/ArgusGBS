<template>
    <div :class="['modal', {fade: fade}]" data-backdrop="static" data-disable="false" data-keyboard="true" :tabindex="tabindex" v-if="closeable">
        <div :class="['modal-dialog', size]">
            <div class="modal-content">
                <div class="modal-header">
                    <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                        <span aria-hidden="true">&times;</span>
                    </button>
                    <h4 class="modal-title text-center text-primary"><span>{{title}}</span></h4>
                </div>
                <form class="form-horizontal" autocomplete="off" ref="form" :enctype="enctype" spellcheck="false">
                    <div class="modal-body">
                        <slot></slot>
                    </div>
                    <div class="modal-footer">
                        <slot name="footer"></slot>
                        <button type="button" class="btn btn-default" data-dismiss="modal">取消</button>
                        <button type="button" class="btn btn-primary" @click.prevent="doSubmit" :disabled="disabled">确定</button>
                    </div>
                </form>
            </div>
        </div>
    </div>
    <div :class="['modal', {fade: fade}]" data-backdrop="static" data-disable="false" data-keyboard="false" :tabindex="tabindex" v-else>
        <div :class="['modal-dialog', size]">
            <div class="modal-content">
                <div class="modal-header">
                    <h4 class="modal-title text-center text-primary"><span>{{title}}</span></h4>
                </div>
                <form class="form-horizontal" autocomplete="off" ref="form" :enctype="enctype" spellcheck="false">
                    <div class="modal-body">
                        <slot></slot>
                    </div>
                    <div class="modal-footer">
                        <slot name="footer"></slot>
                        <button type="button" class="btn btn-primary" @click.prevent="doSubmit" :disabled="disabled">确定</button>
                    </div>
                </form>
            </div>
        </div>
    </div>
</template>

<script>
    import 'jquery-ui/ui/widgets/draggable'
    import $ from 'jquery'

    export default {
        props: {
            enctype: {
                default: ''
            },
            title: {
                default: ''
            },
            size: {
                type: String,
                default: ''
            },
            fade: {
                type: Boolean,
                default: false
            },
            tabindex: {
                default: '-1'
            },
            disabled: {
                type: Boolean,
                default: false
            },
            closeable: {
                type: Boolean,
                default: true,
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
                this.$emit("show");
            }).on("hidden.bs.modal", () => {
                this.errors.clear();
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
            doSubmit() {
                this.$emit("submit");
            }
        }
    }
</script>

<style lang="less" scoped>
    .modal-content {
        overflow: hidden;
    }

    @media screen and(min-width: 992px){
        .modal-dialog.modal-lgg {
            width: 90%;
        }
    }

    @media screen and(min-width: 1200px){
        .modal-dialog.modal-lgg {
            width: 1200px;
        }
    }
</style>
