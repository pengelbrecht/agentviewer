//! Rust Code Test File - Tests syntax highlighting for Rust language features
//! This file demonstrates various Rust language constructs for testing code rendering

use std::collections::{HashMap, HashSet};
use std::error::Error;
use std::fmt::{self, Display, Formatter};
use std::future::Future;
use std::io::{self, Read, Write};
use std::marker::PhantomData;
use std::ops::{Add, Deref, DerefMut};
use std::pin::Pin;
use std::sync::{Arc, Mutex, RwLock};
use std::task::{Context, Poll};

// Constants and statics
const MAX_BUFFER_SIZE: usize = 1024;
const API_VERSION: &str = "1.0.0";
static GLOBAL_COUNTER: AtomicUsize = AtomicUsize::new(0);

use std::sync::atomic::{AtomicUsize, Ordering};

// Enums with variants
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum Status {
    Pending,
    Running { progress: u8 },
    Complete(String),
    Failed { code: i32, message: String },
}

impl Display for Status {
    fn fmt(&self, f: &mut Formatter<'_>) -> fmt::Result {
        match self {
            Status::Pending => write!(f, "Pending"),
            Status::Running { progress } => write!(f, "Running ({}%)", progress),
            Status::Complete(msg) => write!(f, "Complete: {}", msg),
            Status::Failed { code, message } => write!(f, "Failed [{}]: {}", code, message),
        }
    }
}

// Custom error type
#[derive(Debug)]
pub enum AppError {
    NotFound(String),
    InvalidInput { field: String, reason: String },
    NetworkError(io::Error),
    SerializationError(String),
}

impl Display for AppError {
    fn fmt(&self, f: &mut Formatter<'_>) -> fmt::Result {
        match self {
            AppError::NotFound(s) => write!(f, "Not found: {}", s),
            AppError::InvalidInput { field, reason } => {
                write!(f, "Invalid input for '{}': {}", field, reason)
            }
            AppError::NetworkError(e) => write!(f, "Network error: {}", e),
            AppError::SerializationError(s) => write!(f, "Serialization error: {}", s),
        }
    }
}

impl Error for AppError {
    fn source(&self) -> Option<&(dyn Error + 'static)> {
        match self {
            AppError::NetworkError(e) => Some(e),
            _ => None,
        }
    }
}

impl From<io::Error> for AppError {
    fn from(err: io::Error) -> Self {
        AppError::NetworkError(err)
    }
}

// Result type alias
type Result<T> = std::result::Result<T, AppError>;

// Structs with lifetimes
#[derive(Debug, Clone)]
pub struct Config<'a> {
    pub host: &'a str,
    pub port: u16,
    pub timeout_ms: u64,
    pub tags: Vec<String>,
}

impl<'a> Config<'a> {
    pub fn new(host: &'a str) -> Self {
        Config {
            host,
            port: 8080,
            timeout_ms: 30_000,
            tags: Vec::new(),
        }
    }

    pub fn with_port(mut self, port: u16) -> Self {
        self.port = port;
        self
    }

    pub fn with_timeout(mut self, timeout_ms: u64) -> Self {
        self.timeout_ms = timeout_ms;
        self
    }
}

// Generic struct with constraints
pub struct Container<T>
where
    T: Clone + Send + Sync,
{
    items: Vec<T>,
    capacity: usize,
}

impl<T> Container<T>
where
    T: Clone + Send + Sync,
{
    pub fn new(capacity: usize) -> Self {
        Container {
            items: Vec::with_capacity(capacity),
            capacity,
        }
    }

    pub fn push(&mut self, item: T) -> bool {
        if self.items.len() < self.capacity {
            self.items.push(item);
            true
        } else {
            false
        }
    }

    pub fn pop(&mut self) -> Option<T> {
        self.items.pop()
    }

    pub fn len(&self) -> usize {
        self.items.len()
    }

    pub fn is_empty(&self) -> bool {
        self.items.is_empty()
    }
}

// Traits
pub trait Repository<T, ID> {
    fn find(&self, id: ID) -> Option<&T>;
    fn find_all(&self) -> Vec<&T>;
    fn save(&mut self, item: T) -> Result<()>;
    fn delete(&mut self, id: ID) -> Result<()>;
}

pub trait AsyncRepository<T, ID> {
    fn find(&self, id: ID) -> impl Future<Output = Option<T>> + Send;
    fn save(&mut self, item: T) -> impl Future<Output = Result<()>> + Send;
}

// Trait with associated types
pub trait Encoder {
    type Input;
    type Output;
    type Error;

    fn encode(&self, input: Self::Input) -> std::result::Result<Self::Output, Self::Error>;
}

// Newtype pattern with Deref
pub struct UserId(String);

impl Deref for UserId {
    type Target = String;

    fn deref(&self) -> &Self::Target {
        &self.0
    }
}

impl DerefMut for UserId {
    fn deref_mut(&mut self) -> &mut Self::Target {
        &mut self.0
    }
}

impl UserId {
    pub fn new(id: impl Into<String>) -> Self {
        UserId(id.into())
    }
}

// Smart pointer wrapper
pub struct Wrapper<T> {
    inner: Arc<RwLock<T>>,
}

impl<T> Wrapper<T> {
    pub fn new(value: T) -> Self {
        Wrapper {
            inner: Arc::new(RwLock::new(value)),
        }
    }

    pub fn read<F, R>(&self, f: F) -> R
    where
        F: FnOnce(&T) -> R,
    {
        let guard = self.inner.read().unwrap();
        f(&*guard)
    }

    pub fn write<F, R>(&self, f: F) -> R
    where
        F: FnOnce(&mut T) -> R,
    {
        let mut guard = self.inner.write().unwrap();
        f(&mut *guard)
    }
}

impl<T> Clone for Wrapper<T> {
    fn clone(&self) -> Self {
        Wrapper {
            inner: Arc::clone(&self.inner),
        }
    }
}

// Phantom data for type safety
pub struct TypedId<T> {
    id: u64,
    _marker: PhantomData<T>,
}

impl<T> TypedId<T> {
    pub fn new(id: u64) -> Self {
        TypedId {
            id,
            _marker: PhantomData,
        }
    }

    pub fn value(&self) -> u64 {
        self.id
    }
}

// Closures and iterators
fn process_items<T, F, R>(items: Vec<T>, f: F) -> Vec<R>
where
    F: Fn(T) -> R,
{
    items.into_iter().map(f).collect()
}

fn filter_and_transform<T, F, G, R>(items: Vec<T>, filter: F, transform: G) -> Vec<R>
where
    F: Fn(&T) -> bool,
    G: Fn(T) -> R,
{
    items
        .into_iter()
        .filter(filter)
        .map(transform)
        .collect()
}

// Pattern matching examples
fn match_example(value: Option<Result<i32>>) -> String {
    match value {
        Some(Ok(n)) if n > 0 => format!("Positive: {}", n),
        Some(Ok(0)) => "Zero".to_string(),
        Some(Ok(n)) => format!("Negative: {}", n),
        Some(Err(e)) => format!("Error: {}", e),
        None => "None".to_string(),
    }
}

// If let and while let
fn if_let_example(values: &[Option<i32>]) {
    for value in values {
        if let Some(n) = value {
            println!("Got: {}", n);
        }
    }

    let mut iter = values.iter();
    while let Some(Some(n)) = iter.next() {
        println!("While got: {}", n);
    }
}

// Macro definition
macro_rules! create_map {
    ($($key:expr => $value:expr),* $(,)?) => {{
        let mut map = HashMap::new();
        $(
            map.insert($key, $value);
        )*
        map
    }};
}

macro_rules! log {
    ($level:ident, $($arg:tt)*) => {{
        eprintln!(
            "[{}] [{}] {}",
            chrono::Local::now().format("%Y-%m-%d %H:%M:%S"),
            stringify!($level).to_uppercase(),
            format!($($arg)*)
        );
    }};
}

// Async/await
async fn fetch_data(url: &str) -> Result<String> {
    // Simulated async operation
    tokio::time::sleep(std::time::Duration::from_millis(100)).await;
    Ok(format!("Data from {}", url))
}

async fn process_multiple(urls: Vec<&str>) -> Vec<Result<String>> {
    let futures: Vec<_> = urls.into_iter().map(fetch_data).collect();
    futures::future::join_all(futures).await
}

// Custom future implementation
pub struct DelayedValue<T> {
    value: Option<T>,
    delay_ms: u64,
    started: bool,
}

impl<T: Unpin> Future for DelayedValue<T> {
    type Output = T;

    fn poll(mut self: Pin<&mut Self>, _cx: &mut Context<'_>) -> Poll<Self::Output> {
        if !self.started {
            self.started = true;
            Poll::Pending
        } else if let Some(value) = self.value.take() {
            Poll::Ready(value)
        } else {
            Poll::Pending
        }
    }
}

// Operator overloading
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct Point {
    x: f64,
    y: f64,
}

impl Add for Point {
    type Output = Point;

    fn add(self, other: Point) -> Point {
        Point {
            x: self.x + other.x,
            y: self.y + other.y,
        }
    }
}

impl Point {
    pub fn new(x: f64, y: f64) -> Self {
        Point { x, y }
    }

    pub fn distance(&self, other: &Point) -> f64 {
        let dx = self.x - other.x;
        let dy = self.y - other.y;
        (dx * dx + dy * dy).sqrt()
    }
}

// Tests module
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_container() {
        let mut container: Container<i32> = Container::new(10);
        assert!(container.push(1));
        assert!(container.push(2));
        assert_eq!(container.len(), 2);
        assert_eq!(container.pop(), Some(2));
    }

    #[test]
    fn test_point_add() {
        let p1 = Point::new(1.0, 2.0);
        let p2 = Point::new(3.0, 4.0);
        let sum = p1 + p2;
        assert_eq!(sum, Point::new(4.0, 6.0));
    }

    #[test]
    #[should_panic(expected = "assertion failed")]
    fn test_panic() {
        assert!(false);
    }
}

// Main function
fn main() {
    // String types
    let static_str: &'static str = "Hello, World!";
    let string = String::from("Hello, World!");
    let formatted = format!("Value: {}", 42);
    let raw_string = r#"This is a raw string with "quotes""#;
    let byte_string = b"Hello bytes";

    // Numeric literals
    let decimal = 1_000_000;
    let hex = 0xDEAD_BEEF;
    let octal = 0o755;
    let binary = 0b1010_1010;
    let float = 3.14159_f64;

    // Array and slice
    let array: [i32; 5] = [1, 2, 3, 4, 5];
    let slice: &[i32] = &array[1..4];

    // Tuple
    let tuple: (i32, &str, f64) = (42, "hello", 3.14);
    let (a, b, c) = tuple;

    // Vector operations
    let mut vec = vec![1, 2, 3, 4, 5];
    vec.push(6);
    vec.extend([7, 8, 9]);

    // HashMap using macro
    let map = create_map! {
        "one" => 1,
        "two" => 2,
        "three" => 3,
    };

    // Option and Result chaining
    let result: Option<i32> = Some(42)
        .filter(|&n| n > 0)
        .map(|n| n * 2)
        .and_then(|n| if n > 50 { Some(n) } else { None });

    // Error propagation with ?
    fn fallible_operation() -> Result<()> {
        let config = Config::new("localhost");
        println!("Config: {:?}", config);
        Ok(())
    }

    // Closures
    let add = |x: i32, y: i32| x + y;
    let multiply = |x, y| x * y;
    let captured = {
        let multiplier = 10;
        move |x: i32| x * multiplier
    };

    // Iterator chains
    let sum: i32 = (1..=100)
        .filter(|n| n % 2 == 0)
        .map(|n| n * n)
        .take(10)
        .sum();

    // Print values
    println!("Static: {}", static_str);
    println!("String: {}", string);
    println!("Formatted: {}", formatted);
    println!("Raw: {}", raw_string);
    println!("Bytes: {:?}", byte_string);
    println!("Numbers: {} {} {} {} {}", decimal, hex, octal, binary, float);
    println!("Array: {:?}, Slice: {:?}", array, slice);
    println!("Tuple: {:?}, Destructured: {} {} {}", tuple, a, b, c);
    println!("Vector: {:?}", vec);
    println!("Map: {:?}", map);
    println!("Result: {:?}", result);
    println!("Closures: {} {} {}", add(1, 2), multiply(3, 4), captured(5));
    println!("Sum: {}", sum);

    // Run fallible operation
    if let Err(e) = fallible_operation() {
        eprintln!("Error: {}", e);
    }
}
