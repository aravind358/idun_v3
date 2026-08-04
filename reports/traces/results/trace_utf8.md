=== RUN   TestTraceAudit


## Trace: "hi"
2026/08/02 21:00:20 [Kernel] Starting component: Core.Memory (Phase 1)
2026/08/02 21:00:20 [Kernel] Starting component: Core.Scheduler (Phase 1)
2026/08/02 21:00:20 [Kernel] Starting component: Core.Storage (Phase 1)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Boundary (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Bus (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Calibration (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Constitution (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Permission (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Registry (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Infrastructure.Inference (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Infrastructure.Registry (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Workspace (Phase 3)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Attention (Phase 4)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Executive (Phase 4)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Decision (Phase 5)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Planning (Phase 5)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Reasoning (Phase 5)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Understanding (Phase 5)
>>> Understanding V3 WorkspaceBridge Start() is called!
>>> Subscribed to TopicPerception successfully.
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Learning (Phase 6)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Reflection (Phase 6)
2026/08/02 21:00:20 [Kernel] Starting component: Presentation.Router (Phase 6)
2026/08/02 21:00:20 [Kernel] Starting component: World.Service (Phase 6)
2026/08/02 21:00:20 [Kernel] Boot successful
2026/08/02 21:00:20 [Kernel]   Registry  : ServiceRegistry
2026/08/02 21:00:20 [Kernel]   Bus       : CommunicationBus
2026/08/02 21:00:20 [Kernel]   Boundary  : BoundaryEngine
2026/08/02 21:00:20 [Kernel]   Permission: PermissionEngine
[Runtime]
RuntimeHost started.

[World]
Received input:
"hi"

>>> HandlePerception invoked! RawText: "hi"

========== STAGE: user-intent ==========
Envelope ID: interp-b2fffb8531a4bfe1fb6501affc0be789
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "75ca5579-015c-47a6-991d-7c6a0adab528",
  "ParentArtifactID": "b2fffb8531a4bfe1fb6501affc0be789",
  "EnvelopeID": "b2fffb8531a4bfe1fb6501affc0be789",
  "Timestamp": "2026-08-02T21:00:20.5882025+05:30",
  "Status": "UNAMBIGUOUS",
  "PrimaryIntent": "greet_user",
  "CommunicativeAct": "CONVERSATION",
  "PrimaryHypothesis": {
    "Intent": "greet_user",
    "Confidence": 0.98,
    "DeltaFromPrimary": 0,
    "SourceLayer": "Understanding.ReflexiveGrammar",
    "Slots": null
  },
  "AmbiguitySet": null,
  "Topics": null,
  "Entities": null,
  "References": null,
  "TemporalAnchors": null,
  "OpenSlots": null,
  "StatusReason": "",
  "IsConditional": false,
  "ConditionClause": "",
  "ConsequentClause": "",
  "CompoundIntentCount": 1,
  "SecondaryIntents": null,
  "RequiresContext": false,
  "Polarity": {
    "Negated": false,
    "NegationMarker": ""
  },
  "Sentiment": {
    "Valence": "",
    "Intensity": "",
    "Markers": null
  },
  "InputCharacteristics": {
    "ContainsEmoji": false,
    "ContainsSlang": false,
    "ContainsAbbreviation": false,
    "LikelyTypoRecovered": false,
    "Formality": "",
    "InputLength": ""
  },
  "DialoguePosition": {
    "TurnIndex": 0,
    "IsFollowUp": false,
    "IsCorrection": false,
    "IsNewTopic": false
  },
  "ExecutionHints": {
    "AppearsDeterministic": false,
    "AppearsOpenEnded": false,
    "RequiresExternalData": false,
    "RequiresMemoryLookup": false
  },
  "Assumptions": null,
  "Ambiguities": null,
  "MissingInformation": null,
  "Completeness": 1,
  "Confidence": 0.98,
  "ProcessedDurationMs": 0,
  "ValidationTrace": null,
  "ConfidenceTrace": null
}
>>> Reasoning V3 HandleIntent invoked! Envelope ID: interp-b2fffb8531a4bfe1fb6501affc0be789

========== STAGE: active-goals ==========
Envelope ID: f4ef1117-e7f1-4c23-9422-608935354bd2
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "f4ef1117-e7f1-4c23-9422-608935354bd2",
  "ParentArtifactID": "75ca5579-015c-47a6-991d-7c6a0adab528",
  "EnvelopeID": "b2fffb8531a4bfe1fb6501affc0be789",
  "Timestamp": "2026-08-02T21:00:20.5892526+05:30",
  "ResolvedIntent": "greet_user",
  "EnrichedSlots": null,
  "GroundedEntities": null,
  "ResolvedReferences": null,
  "RetrievedContexts": null,
  "ConditionEvaluated": false,
  "ConditionMet": false,
  "TruthEvaluated": false,
  "IsFactuallyTrue": false
}
>>> Planning V3 HandleActiveGoal invoked! Envelope ID: f4ef1117-e7f1-4c23-9422-608935354bd2
>>> Decision V3 HandleCandidatePlan invoked! Envelope ID: 4fde2508-1e07-44e9-b031-ed9c9c6cdb86
>>> Decision V3 SafetyValidator: Checking safety for 'sys-time-1'
>>> Decision V3 PolicyValidator: Checking policy for 'sys-time-1'
>>> Decision V3 AuthValidator: Checking permissions for 'sys-time-1'
>>> Decision V3 BudgetValidator: Checking budget for 'sys-time-1'

========== STAGE: evaluated-options ==========
Envelope ID: 156ffb41-e53d-4530-8a6a-a3c9af175893
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "156ffb41-e53d-4530-8a6a-a3c9af175893",
  "ParentArtifactID": "4fde2508-1e07-44e9-b031-ed9c9c6cdb86",
  "EnvelopeID": "b2fffb8531a4bfe1fb6501affc0be789",
  "Timestamp": "2026-08-02T21:00:20.59295+05:30",
  "Resolution": "APPROVED",
  "Reason": "All evaluations passed unconditionally.",
  "SafetyPassed": true,
  "PolicyPassed": true,
  "BudgetPassed": true,
  "PermissionsPassed": true,
  "Findings": null
}
>>> Executive V3 HandleEvaluatedOption invoked! Envelope ID: 156ffb41-e53d-4530-8a6a-a3c9af175893

========== STAGE: action-execution ==========
Envelope ID: 156ffb41-e53d-4530-8a6a-a3c9af175893
Parsed Payload:
{
  "artifact_id": "792ae4af-383f-41f0-a994-4042377b5db7",
  "parent_artifact_id": "156ffb41-e53d-4530-8a6a-a3c9af175893",
  "envelope_id": "b2fffb8531a4bfe1fb6501affc0be789",
  "timestamp": "2026-08-02T15:30:20.5954887Z",
  "version": "3.0",
  "status": "COMPLETED",
  "node_results": {
    "bc380928-32dd-4e03-a2e7-d061f743304b": {
      "node_id": "bc380928-32dd-4e03-a2e7-d061f743304b",
      "status": "COMPLETED",
      "duration": 0,
      "output_ref": "7bfcf0a79b70192b0cafe652fc3e404a5a18f8bfdb07a4318602ef57603ea3aa"
    }
  },
  "total_duration": 532800
}

========== STAGE: candidate-plans ==========
Envelope ID: 4fde2508-1e07-44e9-b031-ed9c9c6cdb86
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "4fde2508-1e07-44e9-b031-ed9c9c6cdb86",
  "ParentArtifactID": "f4ef1117-e7f1-4c23-9422-608935354bd2",
  "EnvelopeID": "b2fffb8531a4bfe1fb6501affc0be789",
  "Timestamp": "2026-08-02T21:00:20.5913583+05:30",
  "Nodes": [
    {
      "NodeID": "bc380928-32dd-4e03-a2e7-d061f743304b",
      "Capability": "sys-time-1",
      "BoundParams": {}
    }
  ],
  "Edges": null
}
[World]
Published TopicPerception

2026/08/02 21:00:20 [Kernel] Shutdown complete


## Trace: "hello"
2026/08/02 21:00:20 [Kernel] Starting component: Core.Memory (Phase 1)
2026/08/02 21:00:20 [Kernel] Starting component: Core.Scheduler (Phase 1)
2026/08/02 21:00:20 [Kernel] Starting component: Core.Storage (Phase 1)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Boundary (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Bus (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Calibration (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Constitution (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Permission (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Registry (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Infrastructure.Inference (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Infrastructure.Registry (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Workspace (Phase 3)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Attention (Phase 4)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Executive (Phase 4)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Decision (Phase 5)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Planning (Phase 5)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Reasoning (Phase 5)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Understanding (Phase 5)
>>> Understanding V3 WorkspaceBridge Start() is called!
>>> Subscribed to TopicPerception successfully.
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Learning (Phase 6)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Reflection (Phase 6)
2026/08/02 21:00:20 [Kernel] Starting component: Presentation.Router (Phase 6)
2026/08/02 21:00:20 [Kernel] Starting component: World.Service (Phase 6)
2026/08/02 21:00:20 [Kernel] Boot successful
2026/08/02 21:00:20 [Kernel]   Registry  : ServiceRegistry
2026/08/02 21:00:20 [Kernel]   Bus       : CommunicationBus
2026/08/02 21:00:20 [Kernel]   Boundary  : BoundaryEngine
2026/08/02 21:00:20 [Kernel]   Permission: PermissionEngine
[Runtime]
RuntimeHost started.

[World]
Received input:
"hello"

>>> HandlePerception invoked! RawText: "hello"

========== STAGE: user-intent ==========
Envelope ID: interp-a983aaa5ea8cc755e6182bfe332e0882
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "1f381757-8007-47e5-a776-570aeae9dd21",
  "ParentArtifactID": "a983aaa5ea8cc755e6182bfe332e0882",
  "EnvelopeID": "a983aaa5ea8cc755e6182bfe332e0882",
  "Timestamp": "2026-08-02T21:00:20.7002063+05:30",
  "Status": "UNAMBIGUOUS",
  "PrimaryIntent": "greet_user",
  "CommunicativeAct": "CONVERSATION",
  "PrimaryHypothesis": {
    "Intent": "greet_user",
    "Confidence": 0.98,
    "DeltaFromPrimary": 0,
    "SourceLayer": "Understanding.ReflexiveGrammar",
    "Slots": null
  },
  "AmbiguitySet": null,
  "Topics": null,
  "Entities": null,
  "References": null,
  "TemporalAnchors": null,
  "OpenSlots": null,
  "StatusReason": "",
  "IsConditional": false,
  "ConditionClause": "",
  "ConsequentClause": "",
  "CompoundIntentCount": 1,
  "SecondaryIntents": null,
  "RequiresContext": false,
  "Polarity": {
    "Negated": false,
    "NegationMarker": ""
  },
  "Sentiment": {
    "Valence": "",
    "Intensity": "",
    "Markers": null
  },
  "InputCharacteristics": {
    "ContainsEmoji": false,
    "ContainsSlang": false,
    "ContainsAbbreviation": false,
    "LikelyTypoRecovered": false,
    "Formality": "",
    "InputLength": ""
  },
  "DialoguePosition": {
    "TurnIndex": 0,
    "IsFollowUp": false,
    "IsCorrection": false,
    "IsNewTopic": false
  },
  "ExecutionHints": {
    "AppearsDeterministic": false,
    "AppearsOpenEnded": false,
    "RequiresExternalData": false,
    "RequiresMemoryLookup": false
  },
  "Assumptions": null,
  "Ambiguities": null,
  "MissingInformation": null,
  "Completeness": 1,
  "Confidence": 0.98,
  "ProcessedDurationMs": 0,
  "ValidationTrace": null,
  "ConfidenceTrace": null
}
>>> Reasoning V3 HandleIntent invoked! Envelope ID: interp-a983aaa5ea8cc755e6182bfe332e0882

========== STAGE: active-goals ==========
Envelope ID: 1d5b03b3-2e71-499f-ade8-787994b1a92b
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "1d5b03b3-2e71-499f-ade8-787994b1a92b",
  "ParentArtifactID": "1f381757-8007-47e5-a776-570aeae9dd21",
  "EnvelopeID": "a983aaa5ea8cc755e6182bfe332e0882",
  "Timestamp": "2026-08-02T21:00:20.7012842+05:30",
  "ResolvedIntent": "greet_user",
  "EnrichedSlots": null,
  "GroundedEntities": null,
  "ResolvedReferences": null,
  "RetrievedContexts": null,
  "ConditionEvaluated": false,
  "ConditionMet": false,
  "TruthEvaluated": false,
  "IsFactuallyTrue": false
}
>>> Planning V3 HandleActiveGoal invoked! Envelope ID: 1d5b03b3-2e71-499f-ade8-787994b1a92b

========== STAGE: candidate-plans ==========
Envelope ID: fe7114cf-e5bd-41f1-ba8a-c64b0bf09c24
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "fe7114cf-e5bd-41f1-ba8a-c64b0bf09c24",
  "ParentArtifactID": "1d5b03b3-2e71-499f-ade8-787994b1a92b",
  "EnvelopeID": "a983aaa5ea8cc755e6182bfe332e0882",
  "Timestamp": "2026-08-02T21:00:20.7023522+05:30",
  "Nodes": [
    {
      "NodeID": "ab4e4047-c05e-480e-84f0-fc443d7a5a62",
      "Capability": "sys-native-1",
      "BoundParams": {}
    }
  ],
  "Edges": null
}
>>> Decision V3 HandleCandidatePlan invoked! Envelope ID: fe7114cf-e5bd-41f1-ba8a-c64b0bf09c24
>>> Decision V3 SafetyValidator: Checking safety for 'sys-native-1'
>>> Decision V3 PolicyValidator: Checking policy for 'sys-native-1'
>>> Decision V3 AuthValidator: Checking permissions for 'sys-native-1'
>>> Decision V3 BudgetValidator: Checking budget for 'sys-native-1'

========== STAGE: evaluated-options ==========
Envelope ID: a3341181-4326-4fa4-bbdf-8dfb71387750
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "a3341181-4326-4fa4-bbdf-8dfb71387750",
  "ParentArtifactID": "fe7114cf-e5bd-41f1-ba8a-c64b0bf09c24",
  "EnvelopeID": "a983aaa5ea8cc755e6182bfe332e0882",
  "Timestamp": "2026-08-02T21:00:20.7035578+05:30",
  "Resolution": "APPROVED",
  "Reason": "All evaluations passed unconditionally.",
  "SafetyPassed": true,
  "PolicyPassed": true,
  "BudgetPassed": true,
  "PermissionsPassed": true,
  "Findings": null
}
>>> Executive V3 HandleEvaluatedOption invoked! Envelope ID: a3341181-4326-4fa4-bbdf-8dfb71387750

========== STAGE: action-execution ==========
Envelope ID: a3341181-4326-4fa4-bbdf-8dfb71387750
Parsed Payload:
{
  "artifact_id": "57b529c9-379a-4495-ba4d-370bc98c82c0",
  "parent_artifact_id": "a3341181-4326-4fa4-bbdf-8dfb71387750",
  "envelope_id": "a983aaa5ea8cc755e6182bfe332e0882",
  "timestamp": "2026-08-02T15:30:20.7057Z",
  "version": "3.0",
  "status": "COMPLETED",
  "node_results": {
    "ab4e4047-c05e-480e-84f0-fc443d7a5a62": {
      "node_id": "ab4e4047-c05e-480e-84f0-fc443d7a5a62",
      "status": "COMPLETED",
      "duration": 0,
      "output_ref": "7578afec05fe6af9c72c88a8398ae892f1c1210152a5855cdaa33f9ec04054bd"
    }
  },
  "total_duration": 537300
}
[World]
Published TopicPerception

2026/08/02 21:00:20 [Kernel] Shutdown complete


## Trace: "hi bro"
2026/08/02 21:00:20 [Kernel] Starting component: Core.Memory (Phase 1)
2026/08/02 21:00:20 [Kernel] Starting component: Core.Scheduler (Phase 1)
2026/08/02 21:00:20 [Kernel] Starting component: Core.Storage (Phase 1)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Boundary (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Bus (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Calibration (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Constitution (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Permission (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Registry (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Infrastructure.Inference (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Infrastructure.Registry (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Workspace (Phase 3)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Attention (Phase 4)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Executive (Phase 4)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Decision (Phase 5)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Planning (Phase 5)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Reasoning (Phase 5)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Understanding (Phase 5)
>>> Understanding V3 WorkspaceBridge Start() is called!
>>> Subscribed to TopicPerception successfully.
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Learning (Phase 6)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Reflection (Phase 6)
2026/08/02 21:00:20 [Kernel] Starting component: Presentation.Router (Phase 6)
2026/08/02 21:00:20 [Kernel] Starting component: World.Service (Phase 6)
2026/08/02 21:00:20 [Kernel] Boot successful
2026/08/02 21:00:20 [Kernel]   Registry  : ServiceRegistry
2026/08/02 21:00:20 [Kernel]   Bus       : CommunicationBus
2026/08/02 21:00:20 [Kernel]   Boundary  : BoundaryEngine
2026/08/02 21:00:20 [Kernel]   Permission: PermissionEngine
[Runtime]
RuntimeHost started.

[World]
Received input:
"hi bro"

>>> HandlePerception invoked! RawText: "hi bro"

========== STAGE: user-intent ==========
Envelope ID: interp-d6f3767ef56a5246db5aea2173cfc674
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "d89e4614-a421-4e38-a8cf-603281b5cfdc",
  "ParentArtifactID": "d6f3767ef56a5246db5aea2173cfc674",
  "EnvelopeID": "d6f3767ef56a5246db5aea2173cfc674",
  "Timestamp": "2026-08-02T21:00:20.8105068+05:30",
  "Status": "UNAMBIGUOUS",
  "PrimaryIntent": "greet_user",
  "CommunicativeAct": "CONVERSATION",
  "PrimaryHypothesis": {
    "Intent": "greet_user",
    "Confidence": 0.7650000000000001,
    "DeltaFromPrimary": 0,
    "SourceLayer": "Understanding.NeuralClassifier",
    "Slots": null
  },
  "AmbiguitySet": null,
  "Topics": null,
  "Entities": null,
  "References": null,
  "TemporalAnchors": null,
  "OpenSlots": null,
  "StatusReason": "",
  "IsConditional": false,
  "ConditionClause": "",
  "ConsequentClause": "",
  "CompoundIntentCount": 1,
  "SecondaryIntents": null,
  "RequiresContext": false,
  "Polarity": {
    "Negated": false,
    "NegationMarker": ""
  },
  "Sentiment": {
    "Valence": "",
    "Intensity": "",
    "Markers": null
  },
  "InputCharacteristics": {
    "ContainsEmoji": false,
    "ContainsSlang": false,
    "ContainsAbbreviation": false,
    "LikelyTypoRecovered": false,
    "Formality": "",
    "InputLength": ""
  },
  "DialoguePosition": {
    "TurnIndex": 0,
    "IsFollowUp": false,
    "IsCorrection": false,
    "IsNewTopic": false
  },
  "ExecutionHints": {
    "AppearsDeterministic": false,
    "AppearsOpenEnded": false,
    "RequiresExternalData": false,
    "RequiresMemoryLookup": false
  },
  "Assumptions": null,
  "Ambiguities": null,
  "MissingInformation": null,
  "Completeness": 1,
  "Confidence": 0.7650000000000001,
  "ProcessedDurationMs": 0,
  "ValidationTrace": null,
  "ConfidenceTrace": null
}
>>> Reasoning V3 HandleIntent invoked! Envelope ID: interp-d6f3767ef56a5246db5aea2173cfc674

========== STAGE: active-goals ==========
Envelope ID: 86bfae51-1a66-4fbd-b3c7-21513f742fa0
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "86bfae51-1a66-4fbd-b3c7-21513f742fa0",
  "ParentArtifactID": "d89e4614-a421-4e38-a8cf-603281b5cfdc",
  "EnvelopeID": "d6f3767ef56a5246db5aea2173cfc674",
  "Timestamp": "2026-08-02T21:00:20.8115453+05:30",
  "ResolvedIntent": "greet_user",
  "EnrichedSlots": null,
  "GroundedEntities": null,
  "ResolvedReferences": null,
  "RetrievedContexts": null,
  "ConditionEvaluated": false,
  "ConditionMet": false,
  "TruthEvaluated": false,
  "IsFactuallyTrue": false
}
>>> Planning V3 HandleActiveGoal invoked! Envelope ID: 86bfae51-1a66-4fbd-b3c7-21513f742fa0
>>> Decision V3 HandleCandidatePlan invoked! Envelope ID: 9003de31-9981-4120-aa7c-340319530201
>>> Decision V3 SafetyValidator: Checking safety for 'files-native-1'
>>> Decision V3 PolicyValidator: Checking policy for 'files-native-1'
>>> Decision V3 AuthValidator: Checking permissions for 'files-native-1'
>>> Decision V3 BudgetValidator: Checking budget for 'files-native-1'

========== STAGE: evaluated-options ==========
Envelope ID: 89629870-06c5-4324-b502-d685c394659b
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "89629870-06c5-4324-b502-d685c394659b",
  "ParentArtifactID": "9003de31-9981-4120-aa7c-340319530201",
  "EnvelopeID": "d6f3767ef56a5246db5aea2173cfc674",
  "Timestamp": "2026-08-02T21:00:20.8136793+05:30",
  "Resolution": "APPROVED",
  "Reason": "All evaluations passed unconditionally.",
  "SafetyPassed": true,
  "PolicyPassed": true,
  "BudgetPassed": true,
  "PermissionsPassed": true,
  "Findings": null
}
>>> Executive V3 HandleEvaluatedOption invoked! Envelope ID: 89629870-06c5-4324-b502-d685c394659b

========== STAGE: action-execution ==========
Envelope ID: 89629870-06c5-4324-b502-d685c394659b
Parsed Payload:
{
  "artifact_id": "66964718-a512-41cd-865b-cdf12c2077a4",
  "parent_artifact_id": "89629870-06c5-4324-b502-d685c394659b",
  "envelope_id": "d6f3767ef56a5246db5aea2173cfc674",
  "timestamp": "2026-08-02T15:30:20.8152259Z",
  "version": "3.0",
  "status": "COMPLETED",
  "node_results": {
    "cfd4f853-ac65-45d3-82a4-f76f86550c49": {
      "node_id": "cfd4f853-ac65-45d3-82a4-f76f86550c49",
      "status": "COMPLETED",
      "duration": 0,
      "output_ref": "453157c85d98beb8544fca0876e0a8c445220e2d7c108375467f2182dc71e6f3"
    }
  },
  "total_duration": 515200
}

========== STAGE: candidate-plans ==========
Envelope ID: 9003de31-9981-4120-aa7c-340319530201
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "9003de31-9981-4120-aa7c-340319530201",
  "ParentArtifactID": "86bfae51-1a66-4fbd-b3c7-21513f742fa0",
  "EnvelopeID": "d6f3767ef56a5246db5aea2173cfc674",
  "Timestamp": "2026-08-02T21:00:20.81211+05:30",
  "Nodes": [
    {
      "NodeID": "cfd4f853-ac65-45d3-82a4-f76f86550c49",
      "Capability": "files-native-1",
      "BoundParams": {}
    }
  ],
  "Edges": null
}
[World]
Published TopicPerception

2026/08/02 21:00:20 [Kernel] Shutdown complete


## Trace: "bro"
2026/08/02 21:00:20 [Kernel] Starting component: Core.Memory (Phase 1)
2026/08/02 21:00:20 [Kernel] Starting component: Core.Scheduler (Phase 1)
2026/08/02 21:00:20 [Kernel] Starting component: Core.Storage (Phase 1)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Boundary (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Bus (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Calibration (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Constitution (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Permission (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Foundation.Registry (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Infrastructure.Inference (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Infrastructure.Registry (Phase 2)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Workspace (Phase 3)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Attention (Phase 4)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Executive (Phase 4)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Decision (Phase 5)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Planning (Phase 5)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Reasoning (Phase 5)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Understanding (Phase 5)
>>> Understanding V3 WorkspaceBridge Start() is called!
>>> Subscribed to TopicPerception successfully.
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Learning (Phase 6)
2026/08/02 21:00:20 [Kernel] Starting component: Intelligence.Reflection (Phase 6)
2026/08/02 21:00:20 [Kernel] Starting component: Presentation.Router (Phase 6)
2026/08/02 21:00:20 [Kernel] Starting component: World.Service (Phase 6)
2026/08/02 21:00:20 [Kernel] Boot successful
2026/08/02 21:00:20 [Kernel]   Registry  : ServiceRegistry
2026/08/02 21:00:20 [Kernel]   Bus       : CommunicationBus
2026/08/02 21:00:20 [Kernel]   Boundary  : BoundaryEngine
2026/08/02 21:00:20 [Kernel]   Permission: PermissionEngine
[Runtime]
RuntimeHost started.

[World]
Received input:
"bro"

>>> HandlePerception invoked! RawText: "bro"
[Inference]
Request started

[Inference]
Model:
deliberative-parser

[Inference]
Resolved backend:
ollama-local-01

[Inference]
Model:
qwen2.5:1.5b

[Inference]
Sending request...

[Inference]
Ollama prompt len:
3

[Inference]
Ollama prompt preview:
bro

[Inference]
Response received

[Inference]
Execution time:
990.497ms


========== STAGE: user-intent ==========
Envelope ID: interp-2418ade1bdf905e13f72aeaa14b244e5
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "d63eca75-e375-4ee8-b3f9-a2d039137d6e",
  "ParentArtifactID": "2418ade1bdf905e13f72aeaa14b244e5",
  "EnvelopeID": "2418ade1bdf905e13f72aeaa14b244e5",
  "Timestamp": "2026-08-02T21:00:21.9107103+05:30",
  "Status": "FAILED_IMPASSE",
  "PrimaryIntent": "unresolved_intent",
  "CommunicativeAct": "CONVERSATION",
  "PrimaryHypothesis": {
    "Intent": "unresolved_intent",
    "Confidence": 0,
    "DeltaFromPrimary": 0,
    "SourceLayer": "Understanding.ReflexiveGrammar",
    "Slots": null
  },
  "AmbiguitySet": null,
  "Topics": null,
  "Entities": null,
  "References": null,
  "TemporalAnchors": null,
  "OpenSlots": null,
  "StatusReason": "",
  "IsConditional": false,
  "ConditionClause": "",
  "ConsequentClause": "",
  "CompoundIntentCount": 1,
  "SecondaryIntents": null,
  "RequiresContext": false,
  "Polarity": {
    "Negated": false,
    "NegationMarker": ""
  },
  "Sentiment": {
    "Valence": "",
    "Intensity": "",
    "Markers": null
  },
  "InputCharacteristics": {
    "ContainsEmoji": false,
    "ContainsSlang": false,
    "ContainsAbbreviation": false,
    "LikelyTypoRecovered": false,
    "Formality": "",
    "InputLength": ""
  },
  "DialoguePosition": {
    "TurnIndex": 0,
    "IsFollowUp": false,
    "IsCorrection": false,
    "IsNewTopic": false
  },
  "ExecutionHints": {
    "AppearsDeterministic": false,
    "AppearsOpenEnded": false,
    "RequiresExternalData": false,
    "RequiresMemoryLookup": false
  },
  "Assumptions": null,
  "Ambiguities": null,
  "MissingInformation": null,
  "Completeness": 1,
  "Confidence": 0,
  "ProcessedDurationMs": 0,
  "ValidationTrace": null,
  "ConfidenceTrace": null
}
>>> Reasoning V3 HandleIntent invoked! Envelope ID: interp-2418ade1bdf905e13f72aeaa14b244e5
>>> Reasoning V3: Ambiguity/Failure detected (Status: FAILED_IMPASSE). Escalating to DeliberativeSpecialist (S8)...
[Inference]
Request started

[Inference]
Model:
reasoning-deliberative-llm

[Inference]
Resolved backend:
ollama-local-01

[Inference]
Model:
qwen2.5:1.5b

[Inference]
Sending request...

[Inference]
Ollama prompt len:
43

[Inference]
Ollama prompt preview:
Analyze ambiguous intent: unresolved_intent

[Inference]
Ollama HTTP error:
Post "http://127.0.0.1:11434/api/generate": context deadline exceeded

>>> Reasoning V3: DeliberativeSpecialist error: stage S8 deliberative inference failed: inference: execution exceeded budget SLA timeout

========== STAGE: active-goals ==========
Envelope ID: c96fd072-69dd-4574-aaef-7db1a0baaa58
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "c96fd072-69dd-4574-aaef-7db1a0baaa58",
  "ParentArtifactID": "d63eca75-e375-4ee8-b3f9-a2d039137d6e",
  "EnvelopeID": "2418ade1bdf905e13f72aeaa14b244e5",
  "Timestamp": "2026-08-02T21:00:21.9117704+05:30",
  "ResolvedIntent": "unresolved_intent",
  "EnrichedSlots": null,
  "GroundedEntities": null,
  "ResolvedReferences": null,
  "RetrievedContexts": null,
  "ConditionEvaluated": false,
  "ConditionMet": false,
  "TruthEvaluated": false,
  "IsFactuallyTrue": false
}
>>> Planning V3 HandleActiveGoal invoked! Envelope ID: c96fd072-69dd-4574-aaef-7db1a0baaa58
>>> Decision V3 HandleCandidatePlan invoked! Envelope ID: 04a581ef-ed62-4301-8113-0df0512a169b
>>> Decision V3 SafetyValidator: Checking safety for 'app-notes-1'
>>> Decision V3 PolicyValidator: Checking policy for 'app-notes-1'
>>> Decision V3 AuthValidator: Checking permissions for 'app-notes-1'
>>> Decision V3 BudgetValidator: Checking budget for 'app-notes-1'
>>> Executive V3 HandleEvaluatedOption invoked! Envelope ID: 0153c8c3-f916-4ff3-9ca3-35d3eeeeb7b7

========== STAGE: action-execution ==========
Envelope ID: 0153c8c3-f916-4ff3-9ca3-35d3eeeeb7b7
Parsed Payload:
{
  "artifact_id": "0f85fa69-ce5a-4e8b-a870-074dba5058a8",
  "parent_artifact_id": "0153c8c3-f916-4ff3-9ca3-35d3eeeeb7b7",
  "envelope_id": "2418ade1bdf905e13f72aeaa14b244e5",
  "timestamp": "2026-08-02T15:30:31.9303235Z",
  "version": "3.0",
  "status": "COMPLETED",
  "node_results": {
    "f8398096-b859-4a7d-89f3-1365af1b32e6": {
      "node_id": "f8398096-b859-4a7d-89f3-1365af1b32e6",
      "status": "COMPLETED",
      "duration": 0,
      "output_ref": "8c3219dfc57420b86e406c1fa321518415811cf0aff3d2a56dd993d37371cfbf"
    }
  },
  "total_duration": 2248000
}

========== STAGE: evaluated-options ==========
Envelope ID: 0153c8c3-f916-4ff3-9ca3-35d3eeeeb7b7
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "0153c8c3-f916-4ff3-9ca3-35d3eeeeb7b7",
  "ParentArtifactID": "04a581ef-ed62-4301-8113-0df0512a169b",
  "EnvelopeID": "2418ade1bdf905e13f72aeaa14b244e5",
  "Timestamp": "2026-08-02T21:00:31.921912+05:30",
  "Resolution": "APPROVED",
  "Reason": "All evaluations passed unconditionally.",
  "SafetyPassed": true,
  "PolicyPassed": true,
  "BudgetPassed": true,
  "PermissionsPassed": true,
  "Findings": null
}

========== STAGE: candidate-plans ==========
Envelope ID: 04a581ef-ed62-4301-8113-0df0512a169b
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "04a581ef-ed62-4301-8113-0df0512a169b",
  "ParentArtifactID": "c96fd072-69dd-4574-aaef-7db1a0baaa58",
  "EnvelopeID": "2418ade1bdf905e13f72aeaa14b244e5",
  "Timestamp": "2026-08-02T21:00:31.9179045+05:30",
  "Nodes": [
    {
      "NodeID": "f8398096-b859-4a7d-89f3-1365af1b32e6",
      "Capability": "app-notes-1",
      "BoundParams": {}
    }
  ],
  "Edges": null
}
[World]
Published TopicPerception

2026/08/02 21:00:31 [Kernel] Shutdown complete


## Trace: "what time is it"
2026/08/02 21:00:32 [Kernel] Starting component: Core.Memory (Phase 1)
2026/08/02 21:00:32 [Kernel] Starting component: Core.Scheduler (Phase 1)
2026/08/02 21:00:32 [Kernel] Starting component: Core.Storage (Phase 1)
2026/08/02 21:00:32 [Kernel] Starting component: Foundation.Boundary (Phase 2)
2026/08/02 21:00:32 [Kernel] Starting component: Foundation.Bus (Phase 2)
2026/08/02 21:00:32 [Kernel] Starting component: Foundation.Calibration (Phase 2)
2026/08/02 21:00:32 [Kernel] Starting component: Foundation.Constitution (Phase 2)
2026/08/02 21:00:32 [Kernel] Starting component: Foundation.Permission (Phase 2)
2026/08/02 21:00:32 [Kernel] Starting component: Foundation.Registry (Phase 2)
2026/08/02 21:00:32 [Kernel] Starting component: Infrastructure.Inference (Phase 2)
2026/08/02 21:00:32 [Kernel] Starting component: Infrastructure.Registry (Phase 2)
2026/08/02 21:00:32 [Kernel] Starting component: Intelligence.Workspace (Phase 3)
2026/08/02 21:00:32 [Kernel] Starting component: Intelligence.Attention (Phase 4)
2026/08/02 21:00:32 [Kernel] Starting component: Intelligence.Executive (Phase 4)
2026/08/02 21:00:32 [Kernel] Starting component: Intelligence.Decision (Phase 5)
2026/08/02 21:00:32 [Kernel] Starting component: Intelligence.Planning (Phase 5)
2026/08/02 21:00:32 [Kernel] Starting component: Intelligence.Reasoning (Phase 5)
2026/08/02 21:00:32 [Kernel] Starting component: Intelligence.Understanding (Phase 5)
>>> Understanding V3 WorkspaceBridge Start() is called!
>>> Subscribed to TopicPerception successfully.
2026/08/02 21:00:32 [Kernel] Starting component: Intelligence.Learning (Phase 6)
2026/08/02 21:00:32 [Kernel] Starting component: Intelligence.Reflection (Phase 6)
2026/08/02 21:00:32 [Kernel] Starting component: Presentation.Router (Phase 6)
2026/08/02 21:00:32 [Kernel] Starting component: World.Service (Phase 6)
2026/08/02 21:00:32 [Kernel] Boot successful
2026/08/02 21:00:32 [Kernel]   Registry  : ServiceRegistry
2026/08/02 21:00:32 [Kernel]   Bus       : CommunicationBus
2026/08/02 21:00:32 [Kernel]   Boundary  : BoundaryEngine
2026/08/02 21:00:32 [Kernel]   Permission: PermissionEngine
[Runtime]
RuntimeHost started.

[World]
Received input:
"what time is it"

>>> HandlePerception invoked! RawText: "what time is it"
[Inference]
Request started

[Inference]
Model:
deliberative-parser

[Inference]
Resolved backend:
ollama-local-01

[Inference]
Model:
qwen2.5:1.5b

[Inference]
Sending request...

[Inference]
Ollama prompt len:
15

[Inference]
Ollama prompt preview:
what time is it

[Inference]
Response received

[Inference]
Execution time:
2.6905773s


========== STAGE: user-intent ==========
Envelope ID: interp-929dc92185eb31977418f24c67b37a22
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "39bf8586-1165-4a86-bb2e-1d02cbd9bb04",
  "ParentArtifactID": "929dc92185eb31977418f24c67b37a22",
  "EnvelopeID": "929dc92185eb31977418f24c67b37a22",
  "Timestamp": "2026-08-02T21:00:34.7274976+05:30",
  "Status": "FAILED_IMPASSE",
  "PrimaryIntent": "unresolved_intent",
  "CommunicativeAct": "CONVERSATION",
  "PrimaryHypothesis": {
    "Intent": "unresolved_intent",
    "Confidence": 0,
    "DeltaFromPrimary": 0,
    "SourceLayer": "Understanding.ReflexiveGrammar",
    "Slots": null
  },
  "AmbiguitySet": null,
  "Topics": null,
  "Entities": null,
  "References": null,
  "TemporalAnchors": null,
  "OpenSlots": null,
  "StatusReason": "",
  "IsConditional": false,
  "ConditionClause": "",
  "ConsequentClause": "",
  "CompoundIntentCount": 1,
  "SecondaryIntents": null,
  "RequiresContext": false,
  "Polarity": {
    "Negated": false,
    "NegationMarker": ""
  },
  "Sentiment": {
    "Valence": "",
    "Intensity": "",
    "Markers": null
  },
  "InputCharacteristics": {
    "ContainsEmoji": false,
    "ContainsSlang": false,
    "ContainsAbbreviation": false,
    "LikelyTypoRecovered": false,
    "Formality": "",
    "InputLength": ""
  },
  "DialoguePosition": {
    "TurnIndex": 0,
    "IsFollowUp": false,
    "IsCorrection": false,
    "IsNewTopic": false
  },
  "ExecutionHints": {
    "AppearsDeterministic": false,
    "AppearsOpenEnded": false,
    "RequiresExternalData": false,
    "RequiresMemoryLookup": false
  },
  "Assumptions": null,
  "Ambiguities": null,
  "MissingInformation": null,
  "Completeness": 1,
  "Confidence": 0,
  "ProcessedDurationMs": 0,
  "ValidationTrace": null,
  "ConfidenceTrace": null
}
>>> Reasoning V3 HandleIntent invoked! Envelope ID: interp-929dc92185eb31977418f24c67b37a22
>>> Reasoning V3: Ambiguity/Failure detected (Status: FAILED_IMPASSE). Escalating to DeliberativeSpecialist (S8)...
[Inference]
Request started

[Inference]
Model:
reasoning-deliberative-llm

[Inference]
Resolved backend:
ollama-local-01

[Inference]
Model:
qwen2.5:1.5b

[Inference]
Sending request...

[Inference]
Ollama prompt len:
43

[Inference]
Ollama prompt preview:
Analyze ambiguous intent: unresolved_intent

[Inference]
Ollama HTTP error:
Post "http://127.0.0.1:11434/api/generate": context deadline exceeded

>>> Reasoning V3: DeliberativeSpecialist error: stage S8 deliberative inference failed: inference: execution exceeded budget SLA timeout

========== STAGE: active-goals ==========
Envelope ID: 6d147a75-e0c6-47a9-acfd-8e076c75a5e3
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "6d147a75-e0c6-47a9-acfd-8e076c75a5e3",
  "ParentArtifactID": "39bf8586-1165-4a86-bb2e-1d02cbd9bb04",
  "EnvelopeID": "929dc92185eb31977418f24c67b37a22",
  "Timestamp": "2026-08-02T21:00:34.7306393+05:30",
  "ResolvedIntent": "unresolved_intent",
  "EnrichedSlots": null,
  "GroundedEntities": null,
  "ResolvedReferences": null,
  "RetrievedContexts": null,
  "ConditionEvaluated": false,
  "ConditionMet": false,
  "TruthEvaluated": false,
  "IsFactuallyTrue": false
}
>>> Planning V3 HandleActiveGoal invoked! Envelope ID: 6d147a75-e0c6-47a9-acfd-8e076c75a5e3

========== STAGE: candidate-plans ==========
Envelope ID: e98fbd94-9519-48cc-b666-e155e7eb6747
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "e98fbd94-9519-48cc-b666-e155e7eb6747",
  "ParentArtifactID": "6d147a75-e0c6-47a9-acfd-8e076c75a5e3",
  "EnvelopeID": "929dc92185eb31977418f24c67b37a22",
  "Timestamp": "2026-08-02T21:00:44.7350471+05:30",
  "Nodes": [
    {
      "NodeID": "19908d3d-c7d7-4fe9-badc-5e73014e6f87",
      "Capability": "app-calc-1",
      "BoundParams": {}
    }
  ],
  "Edges": null
}
>>> Decision V3 HandleCandidatePlan invoked! Envelope ID: e98fbd94-9519-48cc-b666-e155e7eb6747
>>> Decision V3 SafetyValidator: Checking safety for 'app-calc-1'
>>> Decision V3 PolicyValidator: Checking policy for 'app-calc-1'
>>> Decision V3 AuthValidator: Checking permissions for 'app-calc-1'
>>> Decision V3 BudgetValidator: Checking budget for 'app-calc-1'

========== STAGE: evaluated-options ==========
Envelope ID: 8827ffee-dd7a-4713-8c18-3aaefb90802a
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "8827ffee-dd7a-4713-8c18-3aaefb90802a",
  "ParentArtifactID": "e98fbd94-9519-48cc-b666-e155e7eb6747",
  "EnvelopeID": "929dc92185eb31977418f24c67b37a22",
  "Timestamp": "2026-08-02T21:00:44.7437338+05:30",
  "Resolution": "APPROVED",
  "Reason": "All evaluations passed unconditionally.",
  "SafetyPassed": true,
  "PolicyPassed": true,
  "BudgetPassed": true,
  "PermissionsPassed": true,
  "Findings": null
}
>>> Executive V3 HandleEvaluatedOption invoked! Envelope ID: 8827ffee-dd7a-4713-8c18-3aaefb90802a

========== STAGE: action-execution ==========
Envelope ID: 8827ffee-dd7a-4713-8c18-3aaefb90802a
Parsed Payload:
{
  "artifact_id": "cd6b347b-6fe7-4daa-a9a6-6856dab83162",
  "parent_artifact_id": "8827ffee-dd7a-4713-8c18-3aaefb90802a",
  "envelope_id": "929dc92185eb31977418f24c67b37a22",
  "timestamp": "2026-08-02T15:30:44.7491107Z",
  "version": "3.0",
  "status": "COMPLETED",
  "node_results": {
    "19908d3d-c7d7-4fe9-badc-5e73014e6f87": {
      "node_id": "19908d3d-c7d7-4fe9-badc-5e73014e6f87",
      "status": "COMPLETED",
      "duration": 0,
      "output_ref": "bc4945957d15f94f2dc3723660c0e2cce2050987b306bb767d8a0384522d4121"
    }
  },
  "total_duration": 1634300
}
[World]
Published TopicPerception

2026/08/02 21:00:44 [Kernel] Shutdown complete


## Trace: "3+3"
2026/08/02 21:00:44 [Kernel] Starting component: Core.Memory (Phase 1)
2026/08/02 21:00:44 [Kernel] Starting component: Core.Scheduler (Phase 1)
2026/08/02 21:00:44 [Kernel] Starting component: Core.Storage (Phase 1)
2026/08/02 21:00:44 [Kernel] Starting component: Foundation.Boundary (Phase 2)
2026/08/02 21:00:44 [Kernel] Starting component: Foundation.Bus (Phase 2)
2026/08/02 21:00:44 [Kernel] Starting component: Foundation.Calibration (Phase 2)
2026/08/02 21:00:44 [Kernel] Starting component: Foundation.Constitution (Phase 2)
2026/08/02 21:00:44 [Kernel] Starting component: Foundation.Permission (Phase 2)
2026/08/02 21:00:44 [Kernel] Starting component: Foundation.Registry (Phase 2)
2026/08/02 21:00:44 [Kernel] Starting component: Infrastructure.Inference (Phase 2)
2026/08/02 21:00:44 [Kernel] Starting component: Infrastructure.Registry (Phase 2)
2026/08/02 21:00:44 [Kernel] Starting component: Intelligence.Workspace (Phase 3)
2026/08/02 21:00:44 [Kernel] Starting component: Intelligence.Attention (Phase 4)
2026/08/02 21:00:44 [Kernel] Starting component: Intelligence.Executive (Phase 4)
2026/08/02 21:00:44 [Kernel] Starting component: Intelligence.Decision (Phase 5)
2026/08/02 21:00:44 [Kernel] Starting component: Intelligence.Planning (Phase 5)
2026/08/02 21:00:44 [Kernel] Starting component: Intelligence.Reasoning (Phase 5)
2026/08/02 21:00:44 [Kernel] Starting component: Intelligence.Understanding (Phase 5)
>>> Understanding V3 WorkspaceBridge Start() is called!
>>> Subscribed to TopicPerception successfully.
2026/08/02 21:00:44 [Kernel] Starting component: Intelligence.Learning (Phase 6)
2026/08/02 21:00:44 [Kernel] Starting component: Intelligence.Reflection (Phase 6)
2026/08/02 21:00:44 [Kernel] Starting component: Presentation.Router (Phase 6)
2026/08/02 21:00:44 [Kernel] Starting component: World.Service (Phase 6)
2026/08/02 21:00:44 [Kernel] Boot successful
2026/08/02 21:00:44 [Kernel]   Registry  : ServiceRegistry
2026/08/02 21:00:44 [Kernel]   Bus       : CommunicationBus
2026/08/02 21:00:44 [Kernel]   Boundary  : BoundaryEngine
2026/08/02 21:00:44 [Kernel]   Permission: PermissionEngine
[Runtime]
RuntimeHost started.

[World]
Received input:
"3+3"

>>> HandlePerception invoked! RawText: "3+3"
[Inference]
Request started

[Inference]
Model:
deliberative-parser

[Inference]
Resolved backend:
ollama-local-01

[Inference]
Model:
qwen2.5:1.5b

[Inference]
Sending request...

[Inference]
Ollama prompt len:
3

[Inference]
Ollama prompt preview:
3+3

[Inference]
Response received

[Inference]
Execution time:
1.1629944s


========== STAGE: user-intent ==========
Envelope ID: interp-a9548bf4aeee1dc08c1cdb4b99a322cb
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "7bd053d3-9dee-4cb1-98b3-26f0427b6289",
  "ParentArtifactID": "a9548bf4aeee1dc08c1cdb4b99a322cb",
  "EnvelopeID": "a9548bf4aeee1dc08c1cdb4b99a322cb",
  "Timestamp": "2026-08-02T21:00:46.0195181+05:30",
  "Status": "FAILED_IMPASSE",
  "PrimaryIntent": "unresolved_intent",
  "CommunicativeAct": "CONVERSATION",
  "PrimaryHypothesis": {
    "Intent": "unresolved_intent",
    "Confidence": 0,
    "DeltaFromPrimary": 0,
    "SourceLayer": "Understanding.ReflexiveGrammar",
    "Slots": null
  },
  "AmbiguitySet": null,
  "Topics": null,
  "Entities": null,
  "References": null,
  "TemporalAnchors": null,
  "OpenSlots": null,
  "StatusReason": "",
  "IsConditional": false,
  "ConditionClause": "",
  "ConsequentClause": "",
  "CompoundIntentCount": 1,
  "SecondaryIntents": null,
  "RequiresContext": false,
  "Polarity": {
    "Negated": false,
    "NegationMarker": ""
  },
  "Sentiment": {
    "Valence": "",
    "Intensity": "",
    "Markers": null
  },
  "InputCharacteristics": {
    "ContainsEmoji": false,
    "ContainsSlang": false,
    "ContainsAbbreviation": false,
    "LikelyTypoRecovered": false,
    "Formality": "",
    "InputLength": ""
  },
  "DialoguePosition": {
    "TurnIndex": 0,
    "IsFollowUp": false,
    "IsCorrection": false,
    "IsNewTopic": false
  },
  "ExecutionHints": {
    "AppearsDeterministic": false,
    "AppearsOpenEnded": false,
    "RequiresExternalData": false,
    "RequiresMemoryLookup": false
  },
  "Assumptions": null,
  "Ambiguities": null,
  "MissingInformation": null,
  "Completeness": 1,
  "Confidence": 0,
  "ProcessedDurationMs": 0,
  "ValidationTrace": null,
  "ConfidenceTrace": null
}
>>> Reasoning V3 HandleIntent invoked! Envelope ID: interp-a9548bf4aeee1dc08c1cdb4b99a322cb
>>> Reasoning V3: Ambiguity/Failure detected (Status: FAILED_IMPASSE). Escalating to DeliberativeSpecialist (S8)...
[Inference]
Request started

[Inference]
Model:
reasoning-deliberative-llm

[Inference]
Resolved backend:
ollama-local-01

[Inference]
Model:
qwen2.5:1.5b

[Inference]
Sending request...

[Inference]
Ollama prompt len:
43

[Inference]
Ollama prompt preview:
Analyze ambiguous intent: unresolved_intent

[Inference]
Ollama HTTP error:
Post "http://127.0.0.1:11434/api/generate": context deadline exceeded

>>> Reasoning V3: DeliberativeSpecialist error: stage S8 deliberative inference failed: inference: execution exceeded budget SLA timeout

========== STAGE: active-goals ==========
Envelope ID: 30ce0db1-d04c-444d-97ed-78bca630a72d
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "30ce0db1-d04c-444d-97ed-78bca630a72d",
  "ParentArtifactID": "7bd053d3-9dee-4cb1-98b3-26f0427b6289",
  "EnvelopeID": "a9548bf4aeee1dc08c1cdb4b99a322cb",
  "Timestamp": "2026-08-02T21:00:46.0211645+05:30",
  "ResolvedIntent": "unresolved_intent",
  "EnrichedSlots": null,
  "GroundedEntities": null,
  "ResolvedReferences": null,
  "RetrievedContexts": null,
  "ConditionEvaluated": false,
  "ConditionMet": false,
  "TruthEvaluated": false,
  "IsFactuallyTrue": false
}
>>> Planning V3 HandleActiveGoal invoked! Envelope ID: 30ce0db1-d04c-444d-97ed-78bca630a72d

========== STAGE: candidate-plans ==========
Envelope ID: 43114c79-ee5f-4509-8527-a2f835f891f1
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "43114c79-ee5f-4509-8527-a2f835f891f1",
  "ParentArtifactID": "30ce0db1-d04c-444d-97ed-78bca630a72d",
  "EnvelopeID": "a9548bf4aeee1dc08c1cdb4b99a322cb",
  "Timestamp": "2026-08-02T21:00:56.0261094+05:30",
  "Nodes": [
    {
      "NodeID": "7615bad3-d22b-4566-8a69-03563eeb4975",
      "Capability": "sys-native-1",
      "BoundParams": {}
    }
  ],
  "Edges": null
}
>>> Decision V3 HandleCandidatePlan invoked! Envelope ID: 43114c79-ee5f-4509-8527-a2f835f891f1
>>> Decision V3 SafetyValidator: Checking safety for 'sys-native-1'
>>> Decision V3 PolicyValidator: Checking policy for 'sys-native-1'
>>> Decision V3 AuthValidator: Checking permissions for 'sys-native-1'
>>> Decision V3 BudgetValidator: Checking budget for 'sys-native-1'

========== STAGE: evaluated-options ==========
Envelope ID: c678bacc-2062-4e5f-a51d-b399f95a7d1c
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "c678bacc-2062-4e5f-a51d-b399f95a7d1c",
  "ParentArtifactID": "43114c79-ee5f-4509-8527-a2f835f891f1",
  "EnvelopeID": "a9548bf4aeee1dc08c1cdb4b99a322cb",
  "Timestamp": "2026-08-02T21:00:56.0314084+05:30",
  "Resolution": "APPROVED",
  "Reason": "All evaluations passed unconditionally.",
  "SafetyPassed": true,
  "PolicyPassed": true,
  "BudgetPassed": true,
  "PermissionsPassed": true,
  "Findings": null
}
>>> Executive V3 HandleEvaluatedOption invoked! Envelope ID: c678bacc-2062-4e5f-a51d-b399f95a7d1c

========== STAGE: action-execution ==========
Envelope ID: c678bacc-2062-4e5f-a51d-b399f95a7d1c
Parsed Payload:
{
  "artifact_id": "1eae8aae-c720-4241-822c-ad7ad812d336",
  "parent_artifact_id": "c678bacc-2062-4e5f-a51d-b399f95a7d1c",
  "envelope_id": "a9548bf4aeee1dc08c1cdb4b99a322cb",
  "timestamp": "2026-08-02T15:30:56.0356225Z",
  "version": "3.0",
  "status": "COMPLETED",
  "node_results": {
    "7615bad3-d22b-4566-8a69-03563eeb4975": {
      "node_id": "7615bad3-d22b-4566-8a69-03563eeb4975",
      "status": "COMPLETED",
      "duration": 0,
      "output_ref": "7578afec05fe6af9c72c88a8398ae892f1c1210152a5855cdaa33f9ec04054bd"
    }
  },
  "total_duration": 1060300
}
[World]
Published TopicPerception

2026/08/02 21:00:56 [Kernel] Shutdown complete


## Trace: "tell me a joke"
2026/08/02 21:00:56 [Kernel] Starting component: Core.Memory (Phase 1)
2026/08/02 21:00:56 [Kernel] Starting component: Core.Scheduler (Phase 1)
2026/08/02 21:00:56 [Kernel] Starting component: Core.Storage (Phase 1)
2026/08/02 21:00:56 [Kernel] Starting component: Foundation.Boundary (Phase 2)
2026/08/02 21:00:56 [Kernel] Starting component: Foundation.Bus (Phase 2)
2026/08/02 21:00:56 [Kernel] Starting component: Foundation.Calibration (Phase 2)
2026/08/02 21:00:56 [Kernel] Starting component: Foundation.Constitution (Phase 2)
2026/08/02 21:00:56 [Kernel] Starting component: Foundation.Permission (Phase 2)
2026/08/02 21:00:56 [Kernel] Starting component: Foundation.Registry (Phase 2)
2026/08/02 21:00:56 [Kernel] Starting component: Infrastructure.Inference (Phase 2)
2026/08/02 21:00:56 [Kernel] Starting component: Infrastructure.Registry (Phase 2)
2026/08/02 21:00:56 [Kernel] Starting component: Intelligence.Workspace (Phase 3)
2026/08/02 21:00:56 [Kernel] Starting component: Intelligence.Attention (Phase 4)
2026/08/02 21:00:56 [Kernel] Starting component: Intelligence.Executive (Phase 4)
2026/08/02 21:00:56 [Kernel] Starting component: Intelligence.Decision (Phase 5)
2026/08/02 21:00:56 [Kernel] Starting component: Intelligence.Planning (Phase 5)
2026/08/02 21:00:56 [Kernel] Starting component: Intelligence.Reasoning (Phase 5)
2026/08/02 21:00:56 [Kernel] Starting component: Intelligence.Understanding (Phase 5)
>>> Understanding V3 WorkspaceBridge Start() is called!
>>> Subscribed to TopicPerception successfully.
2026/08/02 21:00:56 [Kernel] Starting component: Intelligence.Learning (Phase 6)
2026/08/02 21:00:56 [Kernel] Starting component: Intelligence.Reflection (Phase 6)
2026/08/02 21:00:56 [Kernel] Starting component: Presentation.Router (Phase 6)
2026/08/02 21:00:56 [Kernel] Starting component: World.Service (Phase 6)
2026/08/02 21:00:56 [Kernel] Boot successful
2026/08/02 21:00:56 [Kernel]   Registry  : ServiceRegistry
2026/08/02 21:00:56 [Kernel]   Bus       : CommunicationBus
2026/08/02 21:00:56 [Kernel]   Boundary  : BoundaryEngine
2026/08/02 21:00:56 [Kernel]   Permission: PermissionEngine
[Runtime]
RuntimeHost started.

[World]
Received input:
"tell me a joke"

>>> HandlePerception invoked! RawText: "tell me a joke"
[Inference]
Request started

[Inference]
Model:
deliberative-parser

[Inference]
Resolved backend:
ollama-local-01

[Inference]
Model:
qwen2.5:1.5b

[Inference]
Sending request...

[Inference]
Ollama prompt len:
14

[Inference]
Ollama prompt preview:
tell me a joke

[Inference]
Response received

[Inference]
Execution time:
1.2187397s


========== STAGE: user-intent ==========
Envelope ID: interp-79363d1d8240738dcc448fe1f4c1a010
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "7d56a8e6-619b-4fec-a02c-b72157f68d37",
  "ParentArtifactID": "79363d1d8240738dcc448fe1f4c1a010",
  "EnvelopeID": "79363d1d8240738dcc448fe1f4c1a010",
  "Timestamp": "2026-08-02T21:00:57.3618553+05:30",
  "Status": "FAILED_IMPASSE",
  "PrimaryIntent": "unresolved_intent",
  "CommunicativeAct": "CONVERSATION",
  "PrimaryHypothesis": {
    "Intent": "unresolved_intent",
    "Confidence": 0,
    "DeltaFromPrimary": 0,
    "SourceLayer": "Understanding.ReflexiveGrammar",
    "Slots": null
  },
  "AmbiguitySet": null,
  "Topics": null,
  "Entities": null,
  "References": null,
  "TemporalAnchors": null,
  "OpenSlots": null,
  "StatusReason": "",
  "IsConditional": false,
  "ConditionClause": "",
  "ConsequentClause": "",
  "CompoundIntentCount": 1,
  "SecondaryIntents": null,
  "RequiresContext": false,
  "Polarity": {
    "Negated": false,
    "NegationMarker": ""
  },
  "Sentiment": {
    "Valence": "",
    "Intensity": "",
    "Markers": null
  },
  "InputCharacteristics": {
    "ContainsEmoji": false,
    "ContainsSlang": false,
    "ContainsAbbreviation": false,
    "LikelyTypoRecovered": false,
    "Formality": "",
    "InputLength": ""
  },
  "DialoguePosition": {
    "TurnIndex": 0,
    "IsFollowUp": false,
    "IsCorrection": false,
    "IsNewTopic": false
  },
  "ExecutionHints": {
    "AppearsDeterministic": false,
    "AppearsOpenEnded": false,
    "RequiresExternalData": false,
    "RequiresMemoryLookup": false
  },
  "Assumptions": null,
  "Ambiguities": null,
  "MissingInformation": null,
  "Completeness": 1,
  "Confidence": 0,
  "ProcessedDurationMs": 0,
  "ValidationTrace": null,
  "ConfidenceTrace": null
}
>>> Reasoning V3 HandleIntent invoked! Envelope ID: interp-79363d1d8240738dcc448fe1f4c1a010
>>> Reasoning V3: Ambiguity/Failure detected (Status: FAILED_IMPASSE). Escalating to DeliberativeSpecialist (S8)...
[Inference]
Request started

[Inference]
Model:
reasoning-deliberative-llm

[Inference]
Resolved backend:
ollama-local-01

[Inference]
Model:
qwen2.5:1.5b

[Inference]
Sending request...

[Inference]
Ollama prompt len:
43

[Inference]
Ollama prompt preview:
Analyze ambiguous intent: unresolved_intent

[Inference]
Ollama HTTP error:
Post "http://127.0.0.1:11434/api/generate": context deadline exceeded

>>> Reasoning V3: DeliberativeSpecialist error: stage S8 deliberative inference failed: inference: execution exceeded budget SLA timeout

========== STAGE: active-goals ==========
Envelope ID: 7bed0779-7a2e-49c1-a389-60e5b9e56bdb
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "7bed0779-7a2e-49c1-a389-60e5b9e56bdb",
  "ParentArtifactID": "7d56a8e6-619b-4fec-a02c-b72157f68d37",
  "EnvelopeID": "79363d1d8240738dcc448fe1f4c1a010",
  "Timestamp": "2026-08-02T21:00:57.3629251+05:30",
  "ResolvedIntent": "unresolved_intent",
  "EnrichedSlots": null,
  "GroundedEntities": null,
  "ResolvedReferences": null,
  "RetrievedContexts": null,
  "ConditionEvaluated": false,
  "ConditionMet": false,
  "TruthEvaluated": false,
  "IsFactuallyTrue": false
}
>>> Planning V3 HandleActiveGoal invoked! Envelope ID: 7bed0779-7a2e-49c1-a389-60e5b9e56bdb

========== STAGE: candidate-plans ==========
Envelope ID: 4ec55a8e-c498-4d9e-88df-a9c967cc6e1a
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "4ec55a8e-c498-4d9e-88df-a9c967cc6e1a",
  "ParentArtifactID": "7bed0779-7a2e-49c1-a389-60e5b9e56bdb",
  "EnvelopeID": "79363d1d8240738dcc448fe1f4c1a010",
  "Timestamp": "2026-08-02T21:01:07.3658665+05:30",
  "Nodes": [
    {
      "NodeID": "2490fffc-7a8f-4dcf-ad5f-96bf5a51776b",
      "Capability": "app-rem-1",
      "BoundParams": {}
    }
  ],
  "Edges": null
}
>>> Decision V3 HandleCandidatePlan invoked! Envelope ID: 4ec55a8e-c498-4d9e-88df-a9c967cc6e1a
>>> Decision V3 SafetyValidator: Checking safety for 'app-rem-1'
>>> Decision V3 PolicyValidator: Checking policy for 'app-rem-1'
>>> Decision V3 AuthValidator: Checking permissions for 'app-rem-1'
>>> Decision V3 BudgetValidator: Checking budget for 'app-rem-1'

========== STAGE: evaluated-options ==========
Envelope ID: f28682f5-ae01-4214-bcfd-f79ab0ae6fa1
Parsed Payload:
{
  "SpecVersion": "3.0",
  "ArtifactID": "f28682f5-ae01-4214-bcfd-f79ab0ae6fa1",
  "ParentArtifactID": "4ec55a8e-c498-4d9e-88df-a9c967cc6e1a",
  "EnvelopeID": "79363d1d8240738dcc448fe1f4c1a010",
  "Timestamp": "2026-08-02T21:01:07.3687976+05:30",
  "Resolution": "APPROVED",
  "Reason": "All evaluations passed unconditionally.",
  "SafetyPassed": true,
  "PolicyPassed": true,
  "BudgetPassed": true,
  "PermissionsPassed": true,
  "Findings": null
}
>>> Executive V3 HandleEvaluatedOption invoked! Envelope ID: f28682f5-ae01-4214-bcfd-f79ab0ae6fa1

========== STAGE: action-execution ==========
Envelope ID: f28682f5-ae01-4214-bcfd-f79ab0ae6fa1
Parsed Payload:
{
  "artifact_id": "f6075fc4-ca1e-430d-93ca-0ccf33290f9d",
  "parent_artifact_id": "f28682f5-ae01-4214-bcfd-f79ab0ae6fa1",
  "envelope_id": "79363d1d8240738dcc448fe1f4c1a010",
  "timestamp": "2026-08-02T15:31:07.3741116Z",
  "version": "3.0",
  "status": "COMPLETED",
  "node_results": {
    "2490fffc-7a8f-4dcf-ad5f-96bf5a51776b": {
      "node_id": "2490fffc-7a8f-4dcf-ad5f-96bf5a51776b",
      "status": "COMPLETED",
      "duration": 0,
      "output_ref": "c605f49e7ee8654e70a462ce16172f53d6ffe501ea51bda8df9f7d1515499667"
    }
  },
  "total_duration": 1134700
}
[World]
Published TopicPerception

2026/08/02 21:01:07 [Kernel] Shutdown complete


## Trace: "create a reminder for tomorrow"
2026/08/02 21:01:07 [Kernel] Starting component: Core.Memory (Phase 1)
2026/08/02 21:01:07 [Kernel] Starting component: Core.Scheduler (Phase 1)
2026/08/02 21:01:07 [Kernel] Starting component: Core.Storage (Phase 1)
2026/08/02 21:01:07 [Kernel] Starting component: Foundation.Boundary (Phase 2)
2026/08/02 21:01:07 [Kernel] Starting component: Foundation.Bus (Phase 2)
2026/08/02 21:01:07 [Kernel] Starting component: Foundation.Calibration (Phase 2)
2026/08/02 21:01:07 [Kernel] Starting component: Foundation.Constitution (Phase 2)
2026/08/02 21:01:07 [Kernel] Starting component: Foundation.Permission (Phase 2)
2026/08/02 21:01:07 [Kernel] Starting component: Foundation.Registry (Phase 2)
2026/08/02 21:01:07 [Kernel] Starting component: Infrastructure.Inference (Phase 2)
2026/08/02 21:01:07 [Kernel] Starting component: Infrastructure.Registry (Phase 2)
2026/08/02 21:01:07 [Kernel] Starting component: Intelligence.Workspace (Phase 3)
2026/08/02 21:01:07 [Kernel] Starting component: Intelligence.Attention (Phase 4)
2026/08/02 21:01:07 [Kernel] Starting component: Intelligence.Executive (Phase 4)
2026/08/02 21:01:07 [Kernel] Starting component: Intelligence.Decision (Phase 5)
2026/08/02 21:01:07 [Kernel] Starting component: Intelligence.Planning (Phase 5)
2026/08/02 21:01:07 [Kernel] Starting component: Intelligence.Reasoning (Phase 5)
2026/08/02 21:01:07 [Kernel] Starting component: Intelligence.Understanding (Phase 5)
>>> Understanding V3 WorkspaceBridge Start() is called!
>>> Subscribed to TopicPerception successfully.
2026/08/02 21:01:07 [Kernel] Starting component: Intelligence.Learning (Phase 6)
2026/08/02 21:01:07 [Kernel] Starting component: Intelligence.Reflection (Phase 6)
2026/08/02 21:01:07 [Kernel] Starting component: Presentation.Router (Phase 6)
2026/08/02 21:01:07 [Kernel] Starting component: World.Service (Phase 6)
2026/08/02 21:01:07 [Kernel] Boot successful
2026/08/02 21:01:07 [Kernel]   Registry  : ServiceRegistry
2026/08/02 21:01:07 [Kernel]   Bus       : CommunicationBus
2026/08/02 21:01:07 [Kernel]   Boundary  : BoundaryEngine
2026/08/02 21:01:07 [Kernel]   Permission: PermissionEngine
[Runtime]
RuntimeHost started.

[World]
Received input:
"create a reminder for tomorrow"

>>> HandlePerception invoked! RawText: "create a reminder for tomorrow"
[Inference]
Request started

[Inference]
Model:
deliberative-parser

[Inference]
Resolved backend:
ollama-local-01

[Inference]
Model:
qwen2.5:1.5b

[Inference]
Sending request...

[Inference]
Ollama prompt len:
30

[Inference]
Ollama prompt preview:
create a reminder for tomorrow

