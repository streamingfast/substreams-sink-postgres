// @generated
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct KvOperations {
    #[prost(message, repeated, tag="1")]
    pub operations: ::prost::alloc::vec::Vec<KvOperation>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct KvOperation {
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
    #[prost(bytes="vec", tag="2")]
    pub value: ::prost::alloc::vec::Vec<u8>,
    #[prost(uint64, tag="3")]
    pub ordinal: u64,
    #[prost(enumeration="kv_operation::Type", tag="4")]
    pub r#type: i32,
}
/// Nested message and enum types in `KVOperation`.
pub mod kv_operation {
    #[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
    #[repr(i32)]
    pub enum Type {
        /// Protobuf default should not be used, this is used so that the consume can ensure that the value was actually specified
        Unset = 0,
        Set = 1,
        Delete = 2,
    }
    impl Type {
        /// String value of the enum field names used in the ProtoBuf definition.
        ///
        /// The values are not transformed in any way and thus are considered stable
        /// (if the ProtoBuf definition does not change) and safe for programmatic use.
        pub fn as_str_name(&self) -> &'static str {
            match self {
                Type::Unset => "UNSET",
                Type::Set => "SET",
                Type::Delete => "DELETE",
            }
        }
        /// Creates an enum from field names used in the ProtoBuf definition.
        pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
            match value {
                "UNSET" => Some(Self::Unset),
                "SET" => Some(Self::Set),
                "DELETE" => Some(Self::Delete),
                _ => None,
            }
        }
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetRequest {
    /// Key to fetch
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetManyRequest {
    /// Keys to fetch
    #[prost(string, repeated, tag="1")]
    pub keys: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetByPrefixRequest {
    /// server may impose a hard limit, trying to go above it would return grpc_error: INVALID_ARGUMENT
    #[prost(uint64, tag="1")]
    pub limit: u64,
    /// requested prefix
    #[prost(string, tag="2")]
    pub prefix: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ScanRequest {
    /// server may impose a hard limit, trying to go above it would return grpc_error: INVALID_ARGUMENT
    #[prost(uint64, tag="1")]
    pub limit: u64,
    /// scanning will start at this point, lexicographically
    #[prost(string, tag="2")]
    pub begin: ::prost::alloc::string::String,
    /// If set, scanning will stop when it reaches this point or above, excluding this exact key
    #[prost(string, optional, tag="3")]
    pub exclusive_end: ::core::option::Option<::prost::alloc::string::String>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetResponse {
    /// Value that was found for the requested key
    #[prost(bytes="vec", tag="1")]
    pub value: ::prost::alloc::vec::Vec<u8>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetManyResponse {
    /// Values that were found for the requested keys
    #[prost(bytes="vec", repeated, tag="1")]
    pub values: ::prost::alloc::vec::Vec<::prost::alloc::vec::Vec<u8>>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetByPrefixResponse {
    /// KV are the key/value pairs that were found with the given prefix
    #[prost(message, repeated, tag="1")]
    pub key_values: ::prost::alloc::vec::Vec<Kv>,
    /// limit_reached is true if there is at least ONE MORE result than the requested limit
    #[prost(bool, tag="2")]
    pub limit_reached: bool,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ScanResponse {
    /// KV are the key/value pairs that were found during scan
    #[prost(message, repeated, tag="1")]
    pub key_values: ::prost::alloc::vec::Vec<Kv>,
    /// limit_reached is true if there is at least ONE MORE result than the requested limit
    #[prost(bool, tag="2")]
    pub limit_reached: bool,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Kv {
    #[prost(string, tag="1")]
    pub key: ::prost::alloc::string::String,
    #[prost(bytes="vec", tag="2")]
    pub value: ::prost::alloc::vec::Vec<u8>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Config {
    #[prost(int64, tag="1")]
    pub start_block: i64,
    #[prost(string, tag="2")]
    pub input_module: ::prost::alloc::string::String,
}
/// This defines a KV Sink to be queried with a generic key access interface (Get, GetMany, Scan, Prefix calls).
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GenericService {
    #[prost(message, optional, tag="1")]
    pub sink_config: ::core::option::Option<Config>,
}
/// This defines configuration to run a WASM query service on top of the KV store being sync'd.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WasmQueryService {
    #[prost(message, optional, tag="1")]
    pub sink_config: ::core::option::Option<Config>,
    /// wasm exports: "kv_get_batch", "kv_get", "kv_scan", "kv_prefix"
    #[prost(bytes="vec", tag="5")]
    pub wasm_query_module: ::prost::alloc::vec::Vec<u8>,
    /// Fully qualified Protobuf Service definition name
    ///
    /// sf.mycustom.v1.MyService
    #[prost(string, tag="2")]
    pub grpc_service: ::prost::alloc::string::String,
}
// @@protoc_insertion_point(module)
