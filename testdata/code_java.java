/**
 * Java Code Test File - Tests syntax highlighting for Java language features
 * This file demonstrates various Java language constructs for testing code rendering
 */
package com.example.testdata;

import java.io.*;
import java.lang.annotation.*;
import java.time.*;
import java.util.*;
import java.util.concurrent.*;
import java.util.function.*;
import java.util.stream.*;

// Custom annotations
@Retention(RetentionPolicy.RUNTIME)
@Target({ElementType.TYPE, ElementType.METHOD})
@interface Component {
    String value() default "";
}

@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.FIELD)
@interface Inject {
}

@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.METHOD)
@Repeatable(Validations.class)
@interface Validate {
    String message();
}

@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.METHOD)
@interface Validations {
    Validate[] value();
}

// Enum with fields and methods
enum Status {
    PENDING("Waiting to start", 0),
    RUNNING("In progress", 1),
    COMPLETE("Finished successfully", 2),
    FAILED("Encountered error", 3);

    private final String description;
    private final int code;

    Status(String description, int code) {
        this.description = description;
        this.code = code;
    }

    public String getDescription() {
        return description;
    }

    public int getCode() {
        return code;
    }

    public static Status fromCode(int code) {
        return Arrays.stream(values())
            .filter(s -> s.code == code)
            .findFirst()
            .orElseThrow(() -> new IllegalArgumentException("Unknown code: " + code));
    }
}

// Record (Java 16+)
record Point(double x, double y) {
    public Point {
        if (Double.isNaN(x) || Double.isNaN(y)) {
            throw new IllegalArgumentException("Coordinates cannot be NaN");
        }
    }

    public double distanceFromOrigin() {
        return Math.sqrt(x * x + y * y);
    }

    public Point translate(double dx, double dy) {
        return new Point(x + dx, y + dy);
    }
}

// Sealed classes (Java 17+)
sealed interface Shape permits Circle, Rectangle, Triangle {
    double area();
    double perimeter();
}

final class Circle implements Shape {
    private final double radius;

    public Circle(double radius) {
        this.radius = radius;
    }

    @Override
    public double area() {
        return Math.PI * radius * radius;
    }

    @Override
    public double perimeter() {
        return 2 * Math.PI * radius;
    }
}

final class Rectangle implements Shape {
    private final double width;
    private final double height;

    public Rectangle(double width, double height) {
        this.width = width;
        this.height = height;
    }

    @Override
    public double area() {
        return width * height;
    }

    @Override
    public double perimeter() {
        return 2 * (width + height);
    }
}

non-sealed class Triangle implements Shape {
    private final double a, b, c;

    public Triangle(double a, double b, double c) {
        this.a = a;
        this.b = b;
        this.c = c;
    }

    @Override
    public double area() {
        double s = (a + b + c) / 2;
        return Math.sqrt(s * (s - a) * (s - b) * (s - c));
    }

    @Override
    public double perimeter() {
        return a + b + c;
    }
}

// Generic interface with bounds
interface Repository<T, ID extends Serializable> {
    Optional<T> findById(ID id);
    List<T> findAll();
    T save(T entity);
    void delete(ID id);
}

// Generic class with multiple type parameters
class Pair<K, V> {
    private final K key;
    private final V value;

    public Pair(K key, V value) {
        this.key = key;
        this.value = value;
    }

    public K getKey() {
        return key;
    }

    public V getValue() {
        return value;
    }

    public <R> R map(BiFunction<K, V, R> mapper) {
        return mapper.apply(key, value);
    }

    @Override
    public String toString() {
        return "Pair(" + key + ", " + value + ")";
    }
}

// Abstract class with generics
abstract class BaseEntity<ID extends Serializable> implements Serializable {
    @Serial
    private static final long serialVersionUID = 1L;

    protected ID id;
    protected LocalDateTime createdAt;
    protected LocalDateTime updatedAt;

    public ID getId() {
        return id;
    }

    public void setId(ID id) {
        this.id = id;
    }

    public abstract boolean validate();
}

// Main class with comprehensive examples
@Component("testRunner")
public class code_java extends BaseEntity<Long> {
    // Constants
    private static final int MAX_RETRIES = 3;
    private static final String DEFAULT_NAME = "Unknown";

    // Static fields
    private static final Map<String, Object> CACHE = new ConcurrentHashMap<>();
    private static volatile int instanceCount = 0;

    // Instance fields with various modifiers
    @Inject
    private String name;
    private final List<String> tags = new ArrayList<>();
    private transient ExecutorService executor;

    // Static initializer block
    static {
        System.out.println("Static initializer");
    }

    // Instance initializer block
    {
        instanceCount++;
        createdAt = LocalDateTime.now();
    }

    // Constructors
    public code_java() {
        this(DEFAULT_NAME);
    }

    public code_java(String name) {
        this.name = name;
        this.executor = Executors.newCachedThreadPool();
    }

    // Getters and setters
    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
        this.updatedAt = LocalDateTime.now();
    }

    @Override
    public boolean validate() {
        return name != null && !name.isBlank();
    }

    // Varargs method
    @Validate(message = "Values cannot be empty")
    @Validate(message = "All values must be positive")
    public int sum(int... values) {
        return Arrays.stream(values).sum();
    }

    // Generic method
    public <T extends Comparable<T>> T findMax(List<T> items) {
        return items.stream()
            .max(Comparable::compareTo)
            .orElseThrow(() -> new NoSuchElementException("Empty list"));
    }

    // Method with functional interface parameters
    public <T, R> List<R> processItems(
            List<T> items,
            Predicate<T> filter,
            Function<T, R> mapper) {
        return items.stream()
            .filter(filter)
            .map(mapper)
            .collect(Collectors.toList());
    }

    // Pattern matching for instanceof (Java 16+)
    public String describe(Object obj) {
        if (obj instanceof String s) {
            return "String of length " + s.length();
        } else if (obj instanceof Integer i) {
            return "Integer: " + i;
        } else if (obj instanceof List<?> list && !list.isEmpty()) {
            return "List with " + list.size() + " elements";
        } else {
            return "Unknown type: " + obj.getClass().getName();
        }
    }

    // Switch expression (Java 14+)
    public String getStatusMessage(Status status) {
        return switch (status) {
            case PENDING -> "Waiting...";
            case RUNNING -> {
                System.out.println("Work in progress");
                yield "Running...";
            }
            case COMPLETE -> "Done!";
            case FAILED -> "Error occurred";
        };
    }

    // Pattern matching in switch (Java 21+)
    public double getArea(Shape shape) {
        return switch (shape) {
            case Circle c -> Math.PI * c.area();
            case Rectangle r -> r.area();
            case Triangle t -> t.area();
        };
    }

    // Try-with-resources
    public String readFile(String path) throws IOException {
        try (var reader = new BufferedReader(new FileReader(path));
             var lines = reader.lines()) {
            return lines.collect(Collectors.joining("\n"));
        }
    }

    // CompletableFuture async operations
    public CompletableFuture<String> fetchAsync(String url) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Thread.sleep(100);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
            return "Response from " + url;
        }, executor);
    }

    // Stream API examples
    public void streamExamples() {
        List<Integer> numbers = List.of(1, 2, 3, 4, 5, 6, 7, 8, 9, 10);

        // Basic stream operations
        var evens = numbers.stream()
            .filter(n -> n % 2 == 0)
            .toList();

        // Mapping and reducing
        var sumOfSquares = numbers.stream()
            .mapToInt(n -> n * n)
            .sum();

        // Collecting to map
        var grouped = numbers.stream()
            .collect(Collectors.groupingBy(
                n -> n % 2 == 0 ? "even" : "odd",
                Collectors.counting()
            ));

        // Parallel stream
        var doubled = numbers.parallelStream()
            .map(n -> n * 2)
            .toList();

        // FlatMap
        List<List<Integer>> nested = List.of(
            List.of(1, 2),
            List.of(3, 4),
            List.of(5, 6)
        );
        var flattened = nested.stream()
            .flatMap(List::stream)
            .toList();

        // Teeing collector (Java 12+)
        var stats = numbers.stream()
            .collect(Collectors.teeing(
                Collectors.summingInt(Integer::intValue),
                Collectors.counting(),
                (sum, count) -> new Pair<>(sum, count)
            ));

        System.out.printf("Evens: %s%n", evens);
        System.out.printf("Sum of squares: %d%n", sumOfSquares);
        System.out.printf("Grouped: %s%n", grouped);
        System.out.printf("Doubled: %s%n", doubled);
        System.out.printf("Flattened: %s%n", flattened);
        System.out.printf("Stats: %s%n", stats);
    }

    // Optional chaining
    public String getUpperName(Optional<code_java> maybeInstance) {
        return maybeInstance
            .map(code_java::getName)
            .map(String::toUpperCase)
            .orElse("N/A");
    }

    // Lambda and method references
    private final Comparator<String> lengthComparator = Comparator
        .comparingInt(String::length)
        .thenComparing(String::compareToIgnoreCase);

    private final Consumer<String> printer = System.out::println;
    private final Supplier<List<String>> listFactory = ArrayList::new;
    private final BiFunction<Integer, Integer, Integer> adder = Integer::sum;

    // Inner class
    class Task implements Runnable {
        private final String taskName;

        public Task(String taskName) {
            this.taskName = taskName;
        }

        @Override
        public void run() {
            System.out.println("Running task: " + taskName);
        }
    }

    // Static nested class
    static class Builder {
        private String name;
        private List<String> tags = new ArrayList<>();

        public Builder name(String name) {
            this.name = name;
            return this;
        }

        public Builder tag(String tag) {
            this.tags.add(tag);
            return this;
        }

        public code_java build() {
            var instance = new code_java(name);
            instance.tags.addAll(tags);
            return instance;
        }
    }

    // Main method
    public static void main(String[] args) {
        // Text blocks (Java 15+)
        String json = """
            {
                "name": "test",
                "values": [1, 2, 3],
                "nested": {
                    "key": "value"
                }
            }
            """;

        // Local variable type inference
        var instance = new code_java.Builder()
            .name("TestInstance")
            .tag("example")
            .tag("demo")
            .build();

        // Anonymous class
        Runnable task = new Runnable() {
            @Override
            public void run() {
                System.out.println("Anonymous task");
            }
        };

        // Lambda expressions
        Runnable lambdaTask = () -> System.out.println("Lambda task");
        Function<String, Integer> length = s -> s.length();
        BiConsumer<String, Integer> biConsumer = (s, i) ->
            System.out.printf("%s: %d%n", s, i);

        // Method references
        List<String> words = List.of("hello", "world", "java");
        words.forEach(System.out::println);

        // Stream with collectors
        String joined = words.stream()
            .map(String::toUpperCase)
            .collect(Collectors.joining(", ", "[", "]"));

        // Date/Time API
        LocalDateTime now = LocalDateTime.now();
        LocalDate date = LocalDate.of(2024, Month.JANUARY, 15);
        Duration duration = Duration.ofHours(2).plusMinutes(30);
        Period period = Period.between(date, LocalDate.now());

        // Try with resources and exception handling
        try {
            instance.streamExamples();
            System.out.println("JSON: " + json);
            System.out.println("Joined: " + joined);
            System.out.println("Now: " + now);
            System.out.println("Duration: " + duration);
            System.out.println("Period: " + period);
        } catch (Exception e) {
            System.err.println("Error: " + e.getMessage());
            e.printStackTrace();
        } finally {
            System.out.println("Cleanup complete");
        }

        // Shutdown executor
        instance.executor.shutdown();
    }
}
