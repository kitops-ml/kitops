<script setup lang="ts">
import { ref, onBeforeUnmount } from 'vue'

import ParameterTooltip from '@/components/ParameterTooltip.vue'
import Accordion from '@/components/ui/Accordion.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Radio from '@/components/ui/Radio.vue'
import Slider from '@/components/ui/Slider.vue'
import Textarea from '@/components/ui/Textarea.vue'
import Toggle from '@/components/ui/Toggle.vue'
import { DEFAULT_PARAMS_VALUES, DEFAULT_SESSION, type Session, type Parameters } from '@/composables/useLlama'
import { SchemaConverter } from '@/services/json-schema-to-grammar'

const props = defineProps<{
  session: Session,
  parameters: Parameters
}>()

const emit = defineEmits<{
  (event: 'close'): void,
  (event: 'save', data: {
    session: Session,
    parameters: Parameters
  }): void
}>()

const session = ref({ ...(props.session || DEFAULT_SESSION) })
const parameters = ref({ ...(props.parameters || DEFAULT_PARAMS_VALUES) })

const isChat = ref(session.value.type === 'chat')
const isModalOpen = ref(true)
const closeModal = () => {
  isModalOpen.value = false
  requestAnimationFrame(() => {
    document.body.classList.remove('overflow-hidden')
    emit('close')
  })
}

const onBackdropClick = () => {
  closeModal()
}

const onModalAnimationEnd = (event: AnimationEvent) => {
  if (!isModalOpen.value || !event.animationName.includes('modalShow')) {
    return
  }

  document.body.classList.add('overflow-hidden')
}

onBeforeUnmount(() => {
  closeModal()
})

const convertJSONSchemaGrammar = async () => {
  try {
    let schema = JSON.parse(parameters.value.grammar)
    const converter = new SchemaConverter({
      prop_order: (parameters.value.prop_order ?? '')
        .split(',')
        .reduce((acc, cur, i) => ({ ...acc, [cur.trim()]: i }), {}),
      allow_fetch: true,
    })
    schema = await converter.resolveRefs(schema, 'input')
    converter.visit(schema, '')
    parameters.value = {
      ...parameters.value,
      grammar: converter.formatGrammar(),
    }
  } catch (e) {
    // @ts-ignore
    alert(`Convert failed: ${e.message}`)
  }
}

const resetToDefaults = () => {
  session.value = { ...DEFAULT_SESSION }
  parameters.value = { ...DEFAULT_PARAMS_VALUES }
  sessionStorage.clear()
}

const save = () => {
  emit('save', {
    session: session.value,
    parameters: parameters.value
  })

  closeModal()
}

const onChatToggle = () => {
  session.value.type = isChat.value
    ? 'chat'
    : 'completion'
}
</script>

<template>
<div
  role="dialog"
  aria-label="Settings modal"
  aria-modal="true"
  class="modal-wrapper"
  :class="{ 'open': isModalOpen }"
  @close="closeModal()">
  <div class="modal-backdrop"
    @click="onBackdropClick">

    <div class="max-h-screen overflow-y-auto">
      <div class="modal-content"
        @click.stop
        @animationend.self="onModalAnimationEnd">
        <div class="flex items-center justify-between">
          <h2 class="text-2xl">Settings</h2>
          <button class="text-off-white hocus:text-gold" @click="closeModal()">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none">
              <path d="M12.0007 10.5865L16.9504 5.63672L18.3646 7.05093L13.4149 12.0007L18.3646 16.9504L16.9504 18.3646L12.0007 13.4149L7.05093 18.3646L5.63672 16.9504L10.5865 12.0007L5.63672 7.05093L7.05093 5.63672L12.0007 10.5865Z" fill="currentColor" />
            </svg>
          </button>
        </div>

        <label class="flex items-center gap-2 text-xl mt-10">
          Chat <Toggle v-model="isChat" @change="onChatToggle" />
        </label>

        <template v-if="isChat">
          <div class="flex gap-6 mt-10">
            <Input v-model="session.user" placeholder="eg. User" wrapper-class="flex-1">
              <template #label="{ className }">
                <ParameterTooltip
                  :className
                  message="The name of the user interacting with the chatbot.">
                  User name
                </ParameterTooltip>
              </template>
            </Input>

            <Input v-model="session.char" placeholder="eg. Llama" wrapper-class="flex-1">
              <template #label="{ className }">
                <ParameterTooltip
                  :className
                  message="The name of the chatbot.">
                  Bot name
                </ParameterTooltip>
              </template>
            </Input>
          </div>
        </template>

        <div class="mt-10">
          <Textarea
            autogrow
            id="textarea-prompt"
            placeholder="Prompt"
            rows="3"
            v-model="session.prompt"
            class="h-28">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="Specific guidelines that involve providing an input sequence of tokens to generate a continuation or completion of the sequence.">
                Prompt
              </ParameterTooltip>
            </template>
          </Textarea>
        </div>

        <Accordion id="accordion-templates" summary-class="border-b border-b-elevation-05 py-2 text-xl mt-22" content-class="space-y-6 mt-10">
          <template #title>Templates</template>

          <Textarea
            id="textarea-templates"
            class="h-36"
            v-model="session.template">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="The template used to generate the initial prompt for the chatbot.">
                Prompt template
              </ParameterTooltip>
            </template>
          </Textarea>

          <Input id="Message" placeholder="Chat history template" model-value="{{ name }}: {{ message }}" wrapper-class="!mb-10">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="The template used to format the chat history for display to the user.">
                Chat history template
              </ParameterTooltip>
            </template>
          </Input>
        </Accordion>

        <Accordion id="accordion-options" summary-class="border-b border-b-elevation-05 py-2 text-xl mt-12" content-class="md:grid grid-cols-2 md:gap-6 xs:space-y-6 mt-10">
          <template #title>Text Generation Controls</template>

          <Slider v-model.number="parameters.temperature" :min="0" :step="0.01" :max="2">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="Adjust the randomness of the output where higher temperatures result in more random outputs and lower temperatures result in more deterministic outputs.">
                Temperature
              </ParameterTooltip>
            </template>
          </Slider>

          <Slider v-model.number="parameters.n_predict" :min="-1" :step="1" :max="2048">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="Number of tokens that the model should generate in response">
                Predictions
              </ParameterTooltip>
            </template>
          </Slider>
          <Slider v-model.number="parameters.top_k" :max="100" :min="-1">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="Selecting the top K most probable tokens from a probability distribution and normalizing their probabilities to create a new distribution for sampling, where K is a user-defined parameter.">
                Top-K sampling
              </ParameterTooltip>
            </template>
          </Slider>
          <Slider v-model.number="parameters.top_p" :max="1" :min="0" :step="0.01">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="Selecting the smallest number of top tokens whose cumulative probability is at least p, where p is a user-defined parameter, and sampling from those tokens after normalizing their probabilities.">
                Top-P sampling
              </ParameterTooltip>
            </template>
          </Slider>
          <Slider v-model.number="parameters.min_p" :max="1" :min="0" :step="0.01" wrapper-class="!mb-10">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="Sets a minimum probability threshold for tokens to be considered, based on the confidence of the highest probability token, allowing for more diverse choices while preventing the model from considering too many or too few tokens.">
                Min-P sampling
              </ParameterTooltip>
            </template>
          </Slider>
        </Accordion>

        <Accordion id="accordion-sample-diversity" summary-class="border-b border-b-elevation-05 py-2 text-xl mt-12" content-class="md:grid grid-cols-2 md:gap-6 xs:space-y-6 mt-10">
          <template #title>Sampling and Diversity</template>

          <Slider v-model.number="parameters.tfs_z" :max="1" :min="0" :step="0.01">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="Modifies probability distribution to carefully cut off least likely tokens.">
                TFS-Z
              </ParameterTooltip>
            </template>
          </Slider>
          <Slider v-model.number="parameters.typical_p" :max="1" :min="0" :step="0.01">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="Enable locally typical sampling with parameter p">
                Typical P
              </ParameterTooltip>
            </template>
          </Slider>
          <Slider v-model.number="parameters.presence_penalty" :max="1" :min="0" :step="0.01">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="The penalty to apply to tokens based on their presence in the prompt.">
                Presence penalty
              </ParameterTooltip>
            </template>
          </Slider>
          <Slider v-model.number="parameters.frequency_penalty" :max="1" :min="0" :step="0.01">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="The penalty to apply to tokens based on their frequency in the prompt.">
                Frequency penalty
              </ParameterTooltip>
            </template>
          </Slider>


          <hr class="col-span-2 border-elevation-05 my-8">

          <div class="col-span-2 flex flex-col md:flex-row items-center gap-16">
            <Radio label="no Mirostat" name="mirostat" :value="0" v-model.number="parameters.mirostat">
              <template #label="{ className }">
                <ParameterTooltip
                  :className
                  message="Mirostat is not used">
                  No Mirostat
                </ParameterTooltip>
              </template>
            </Radio>

            <Radio label="Mirostat v1" name="mirostat" :value="1" v-model.number="parameters.mirostat">
              <template #label="{ className }">
                <ParameterTooltip
                  :className
                  message="Adjusts the value of k in top-k decoding to keep the perplexity within a specific range. (Top K, Nucleus, Tail Free and Locally Typical samplers are ignored if used.)">
                  Mirostat v1
                </ParameterTooltip>
              </template>
            </Radio>

            <Radio label="Mirostat v2" name="mirostat" :value="2" v-model.number="parameters.mirostat">
              <template #label="{ className }">
                <ParameterTooltip
                  :className
                  message="Adjusts the value of k in top-k decoding to keep the perplexity within a specific range. (Top K, Nucleus, Tail Free and Locally Typical samplers are ignored if used.)">
                  Mirostat v2
                </ParameterTooltip>
              </template>
            </Radio>

          </div>

          <hr class="col-span-2 border-elevation-05 my-8">

          <Slider v-model.number="parameters.mirostat_tau" :max="10" :min="0" :step="0.01">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="The target cross-entropy (or surprise) value you want to achieve for the generated text. A higher value corresponds to more surprising or less predictable text, while a lower value corresponds to less surprising or more predictable text.">
                Mirostat tau
              </ParameterTooltip>
            </template>
          </Slider>

          <Slider v-model.number="parameters.mirostat_eta" :max="1" :min="0" :step="0.01">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="The learning rate used to update `mu` based on the error between the target and observed surprisal of the sampled word. A larger learning rate will cause `mu` to be updated more quickly, while a smaller learning rate will result in slower updates.">
                Mirostat eta
              </ParameterTooltip>
            </template>
          </Slider>

          <Slider v-model.number="parameters.repeat_penalty" :min="0" :step="0.01" :max="2">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="The penalty to apply to repeated tokens.">
                Penalize repeat sequence
              </ParameterTooltip>
            </template>
          </Slider>
          <Slider v-model.number="parameters.repeat_last_n" :min="0" :max="2048">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="Number of tokens that should be considered for repetition penalty.">
                Consider N tokens for penalize
              </ParameterTooltip>
            </template>
          </Slider>
        </Accordion>

        <Accordion id="accordion-advanced-settings" summary-class="border-b border-b-elevation-05 py-2 text-xl mt-12" content-class="md:grid grid-cols-2 md:gap-6 xs:space-y-6 mt-10">
          <template #title>Advanced Settings and Customization</template>

          <div class="col-span-2">
            <Textarea id="grammar" placeholder="Use gbnf or JSON Schema+convert" wrapper-class="mt-8" v-model="parameters.grammar">
              <template #label="{ className }">
                <ParameterTooltip
                  :className
                  message="A grammar to use for constrained sampling.">
                  Grammar
                </ParameterTooltip>
              </template>
            </Textarea>

            <div class="flex gap-6 mt-6 mb-10">
              <Input id="order" wrapper-class="flex-1" placeholder="order: prop1, prop2, prop3" v-model="parameters.prop_order" />
              <Button secondary :disabled="!parameters.grammar" @click="convertJSONSchemaGrammar">CONVERT JSON SCHEMA</Button>
            </div>

            <Input id="api-key" placeholder="Enter API Key" class="!w-1/2" v-model="parameters.api_key">
              <template #label="{ className }">
                <ParameterTooltip
                  :className
                  message="Unique API Key">
                  API Key
                </ParameterTooltip>
              </template>
            </Input>
          </div>
        </Accordion>

        <Accordion id="accordion-probability-stats" summary-class="border-b border-b-elevation-05 py-2 text-xl mt-12" content-class="md:grid grid-cols-2 md:gap-6 xs:space-y-6 mt-10">
          <template #title>Probability and Statistical Controls</template>

          <Slider v-model.number="parameters.n_probs">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="Display the top N most probable tokens along with their probabilities.">
                Show probabilities
              </ParameterTooltip>
            </template>
          </Slider>

          <Slider v-model.number="parameters.min_keep">
            <template #label="{ className }">
              <ParameterTooltip
                :className
                message="Controls the minimum probability threshold for each sampler.">
                Min probabilities from each sampler
              </ParameterTooltip>
            </template>
          </Slider>
        </Accordion>

        <div class="mt-16 flex justify-between items-center">
          <Button @click="resetToDefaults" class="!border-transparent !bg-transparent hover:!bg-transparent hover:!text-cornflower !text-white !px-0">RESET TO DEFAULT SETTINGS</Button>

          <div>
            <Button secondary class="!border-transparent !bg-transparent mr-6" @click="closeModal">CANCEL</Button>
            <Button primary @click="save">SAVE</Button>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>
</template>

<style scoped>
/* Animations */
@keyframes fadeIn {
  from {
    opacity: 0;
  }

  to {
    opacity: 1;
  }
}

@keyframes modalShow {
  from {
    transform: scale(0.95) translateY(-2.5%);
  }

  to {
    transform: scale(1);
  }
}

.modal-wrapper {
  @apply fixed left-0 top-0 w-screen h-screen z-10 table;
}

.modal-wrapper .modal-backdrop {
  @apply bg-black bg-opacity-50 table-cell text-center align-bottom md:align-middle;
}

.modal-wrapper .modal-content {
  @apply bg-elevation-01 w-[95%] max-w-3xl relative block text-left mx-auto my-22;
  @apply p-12;
}

.modal-wrapper.open .modal-backdrop {
  animation: fadeIn 150ms ease-in-out;
}

.modal-wrapper.open .modal-content {
  animation: modalShow 150ms ease-in-out forwards;
}
</style>
<!-- AGENT_MODIFIED: Human review required before merge -->
