/**
 * C++ Code Test File - Tests syntax highlighting for C++ language features
 * This file demonstrates various C++ language constructs for testing code rendering
 */

#include <algorithm>
#include <array>
#include <chrono>
#include <concepts>
#include <coroutine>
#include <expected>
#include <format>
#include <functional>
#include <iostream>
#include <map>
#include <memory>
#include <mutex>
#include <optional>
#include <ranges>
#include <span>
#include <string>
#include <string_view>
#include <thread>
#include <tuple>
#include <variant>
#include <vector>

// Namespace
namespace agentviewer::testing {

// Constants
constexpr int MAX_BUFFER_SIZE = 1024;
constexpr std::string_view API_VERSION = "1.0.0";

// Enum class (strongly typed)
enum class Status {
    Pending,
    Running,
    Complete,
    Failed
};

// Concepts (C++20)
template<typename T>
concept Printable = requires(T t) {
    { std::cout << t } -> std::same_as<std::ostream&>;
};

template<typename T>
concept Comparable = requires(T a, T b) {
    { a < b } -> std::convertible_to<bool>;
    { a == b } -> std::convertible_to<bool>;
};

template<typename T>
concept Container = requires(T c) {
    typename T::value_type;
    typename T::iterator;
    { c.begin() } -> std::same_as<typename T::iterator>;
    { c.end() } -> std::same_as<typename T::iterator>;
    { c.size() } -> std::convertible_to<std::size_t>;
};

// Template class with concept constraint
template<Comparable T>
class OrderedSet {
public:
    void insert(T value) {
        auto it = std::lower_bound(data_.begin(), data_.end(), value);
        if (it == data_.end() || *it != value) {
            data_.insert(it, std::move(value));
        }
    }

    bool contains(const T& value) const {
        return std::binary_search(data_.begin(), data_.end(), value);
    }

    [[nodiscard]] std::size_t size() const noexcept { return data_.size(); }
    [[nodiscard]] bool empty() const noexcept { return data_.empty(); }

    auto begin() { return data_.begin(); }
    auto end() { return data_.end(); }
    auto begin() const { return data_.cbegin(); }
    auto end() const { return data_.cend(); }

private:
    std::vector<T> data_;
};

// RAII wrapper for resources
template<typename T, typename Deleter = std::default_delete<T>>
class UniqueResource {
public:
    explicit UniqueResource(T* ptr = nullptr, Deleter deleter = Deleter())
        : ptr_(ptr), deleter_(std::move(deleter)) {}

    ~UniqueResource() {
        if (ptr_) {
            deleter_(ptr_);
        }
    }

    // Delete copy operations
    UniqueResource(const UniqueResource&) = delete;
    UniqueResource& operator=(const UniqueResource&) = delete;

    // Move operations
    UniqueResource(UniqueResource&& other) noexcept
        : ptr_(std::exchange(other.ptr_, nullptr)),
          deleter_(std::move(other.deleter_)) {}

    UniqueResource& operator=(UniqueResource&& other) noexcept {
        if (this != &other) {
            reset();
            ptr_ = std::exchange(other.ptr_, nullptr);
            deleter_ = std::move(other.deleter_);
        }
        return *this;
    }

    T* get() const noexcept { return ptr_; }
    T& operator*() const { return *ptr_; }
    T* operator->() const noexcept { return ptr_; }
    explicit operator bool() const noexcept { return ptr_ != nullptr; }

    void reset(T* ptr = nullptr) {
        if (ptr_) {
            deleter_(ptr_);
        }
        ptr_ = ptr;
    }

    T* release() noexcept {
        return std::exchange(ptr_, nullptr);
    }

private:
    T* ptr_;
    [[no_unique_address]] Deleter deleter_;
};

// Variadic template
template<typename... Args>
void log(std::string_view format, Args&&... args) {
    auto now = std::chrono::system_clock::now();
    std::cout << std::format("[{}] {}\n",
        std::chrono::floor<std::chrono::seconds>(now),
        std::vformat(format, std::make_format_args(args...)));
}

// Fold expression
template<typename... Args>
auto sum(Args... args) {
    return (args + ...);
}

template<typename... Args>
void print_all(Args&&... args) {
    (std::cout << ... << std::forward<Args>(args)) << '\n';
}

// CRTP (Curiously Recurring Template Pattern)
template<typename Derived>
class Cloneable {
public:
    std::unique_ptr<Derived> clone() const {
        return std::make_unique<Derived>(static_cast<const Derived&>(*this));
    }
};

// Abstract base class
class Entity {
public:
    Entity() = default;
    virtual ~Entity() = default;

    // Rule of five
    Entity(const Entity&) = default;
    Entity(Entity&&) noexcept = default;
    Entity& operator=(const Entity&) = default;
    Entity& operator=(Entity&&) noexcept = default;

    virtual void update() = 0;
    virtual void render() const = 0;

protected:
    int id_ = 0;
    std::string name_;
};

// Derived class with override
class Player : public Entity, public Cloneable<Player> {
public:
    Player(std::string name, int health = 100)
        : name_(std::move(name)), health_(health) {}

    void update() override {
        // Update logic
    }

    void render() const override {
        std::cout << std::format("Player: {} (HP: {})\n", name_, health_);
    }

    void takeDamage(int damage) noexcept {
        health_ = std::max(0, health_ - damage);
    }

    [[nodiscard]] bool isAlive() const noexcept { return health_ > 0; }

private:
    std::string name_;
    int health_;
};

// Aggregate initialization (C++20)
struct Config {
    std::string host = "localhost";
    int port = 8080;
    bool debug = false;
    std::vector<std::string> tags = {};

    auto operator<=>(const Config&) const = default;
};

// Structured bindings and tuple
auto get_user_info() -> std::tuple<std::string, int, bool> {
    return {"Alice", 30, true};
}

// std::variant and std::visit
using Value = std::variant<int, double, std::string, std::vector<int>>;

struct ValuePrinter {
    void operator()(int i) const { std::cout << "int: " << i << '\n'; }
    void operator()(double d) const { std::cout << "double: " << d << '\n'; }
    void operator()(const std::string& s) const { std::cout << "string: " << s << '\n'; }
    void operator()(const std::vector<int>& v) const {
        std::cout << "vector: [";
        for (auto it = v.begin(); it != v.end(); ++it) {
            if (it != v.begin()) std::cout << ", ";
            std::cout << *it;
        }
        std::cout << "]\n";
    }
};

// Lambda expressions
inline constexpr auto square = [](auto x) { return x * x; };
inline constexpr auto add = [](auto a, auto b) { return a + b; };

// Generic lambda with explicit template parameter (C++20)
inline constexpr auto make_pair = []<typename T, typename U>(T t, U u) {
    return std::pair{std::move(t), std::move(u)};
};

// Range-based algorithms (C++20)
void ranges_examples() {
    std::vector<int> numbers = {1, 2, 3, 4, 5, 6, 7, 8, 9, 10};

    // Views
    auto evens = numbers | std::views::filter([](int n) { return n % 2 == 0; });
    auto squared = evens | std::views::transform(square);

    // Take and drop
    auto first_three = numbers | std::views::take(3);
    auto skip_two = numbers | std::views::drop(2);

    // Reverse
    auto reversed = numbers | std::views::reverse;

    // Enumerate (C++23)
    // for (auto [idx, val] : numbers | std::views::enumerate) {
    //     std::cout << idx << ": " << val << '\n';
    // }

    // Collect to vector
    std::vector<int> result;
    std::ranges::copy(squared, std::back_inserter(result));

    // Ranges algorithms
    auto sum = std::ranges::fold_left(numbers, 0, std::plus{});
    auto max = std::ranges::max(numbers);
    auto min = std::ranges::min(numbers);

    std::cout << std::format("Sum: {}, Max: {}, Min: {}\n", sum, max, min);
}

// Coroutine (C++20)
struct Task {
    struct promise_type {
        Task get_return_object() { return {}; }
        std::suspend_never initial_suspend() noexcept { return {}; }
        std::suspend_never final_suspend() noexcept { return {}; }
        void return_void() {}
        void unhandled_exception() { std::terminate(); }
    };
};

Task example_coroutine() {
    std::cout << "Start coroutine\n";
    co_return;
}

// Thread-safe singleton
template<typename T>
class Singleton {
public:
    static T& instance() {
        static T instance;
        return instance;
    }

    Singleton(const Singleton&) = delete;
    Singleton& operator=(const Singleton&) = delete;

protected:
    Singleton() = default;
    ~Singleton() = default;
};

// Thread-safe queue
template<typename T>
class ThreadSafeQueue {
public:
    void push(T value) {
        std::lock_guard lock(mutex_);
        queue_.push_back(std::move(value));
        cv_.notify_one();
    }

    std::optional<T> try_pop() {
        std::lock_guard lock(mutex_);
        if (queue_.empty()) {
            return std::nullopt;
        }
        T value = std::move(queue_.front());
        queue_.pop_front();
        return value;
    }

    T wait_and_pop() {
        std::unique_lock lock(mutex_);
        cv_.wait(lock, [this] { return !queue_.empty(); });
        T value = std::move(queue_.front());
        queue_.pop_front();
        return value;
    }

    [[nodiscard]] bool empty() const {
        std::lock_guard lock(mutex_);
        return queue_.empty();
    }

private:
    mutable std::mutex mutex_;
    std::condition_variable cv_;
    std::deque<T> queue_;
};

// Span example (C++20)
void process_span(std::span<const int> data) {
    for (int value : data) {
        std::cout << value << ' ';
    }
    std::cout << '\n';
}

// Expected (C++23)
enum class Error {
    InvalidInput,
    NotFound,
    Timeout
};

std::expected<int, Error> parse_int(std::string_view str) {
    try {
        return std::stoi(std::string(str));
    } catch (...) {
        return std::unexpected(Error::InvalidInput);
    }
}

} // namespace agentviewer::testing

// Main function
int main() {
    using namespace agentviewer::testing;

    // Raw string literal
    const char* raw = R"(
        This is a raw string
        with "quotes" and \backslashes
        spanning multiple lines
    )";

    // User-defined literals
    using namespace std::literals;
    auto str = "Hello"s;
    auto dur = 100ms;
    auto view = "World"sv;

    // Structured bindings
    auto [name, age, active] = get_user_info();
    std::cout << std::format("Name: {}, Age: {}, Active: {}\n", name, age, active);

    // Aggregate initialization
    Config config{
        .host = "api.example.com",
        .port = 443,
        .debug = true,
        .tags = {"prod", "v2"}
    };

    // Optional
    std::optional<int> maybeValue = 42;
    auto value = maybeValue.value_or(0);

    // Variant
    Value v = std::vector<int>{1, 2, 3};
    std::visit(ValuePrinter{}, v);

    // Smart pointers
    auto unique = std::make_unique<Player>("Hero", 100);
    auto shared = std::make_shared<Player>("Sidekick", 80);
    std::weak_ptr<Player> weak = shared;

    // Lambda captures
    int multiplier = 10;
    auto multiply = [multiplier](int x) { return x * multiplier; };
    auto by_ref = [&multiplier](int x) { multiplier = x; return multiplier; };
    auto by_move = [str = std::move(str)]() { return str; };

    // constexpr if
    constexpr auto check = []<typename T>(T value) {
        if constexpr (std::is_integral_v<T>) {
            return value * 2;
        } else if constexpr (std::is_floating_point_v<T>) {
            return value / 2.0;
        } else {
            return value;
        }
    };

    // Array and span
    std::array<int, 5> arr = {1, 2, 3, 4, 5};
    process_span(arr);

    // Ranges
    ranges_examples();

    // Expected
    auto result = parse_int("42");
    if (result) {
        std::cout << "Parsed: " << *result << '\n';
    } else {
        std::cout << "Parse error\n";
    }

    // Threading
    ThreadSafeQueue<int> queue;
    std::jthread producer([&queue](std::stop_token token) {
        for (int i = 0; i < 5 && !token.stop_requested(); ++i) {
            queue.push(i);
            std::this_thread::sleep_for(10ms);
        }
    });

    std::this_thread::sleep_for(100ms);

    while (!queue.empty()) {
        if (auto val = queue.try_pop()) {
            std::cout << "Got: " << *val << '\n';
        }
    }

    // Output
    log("Application finished with value: {}", value);
    std::cout << raw << '\n';
    std::cout << std::format("Multiply: {}, Check: {}\n", multiply(5), check(42));

    return 0;
}
